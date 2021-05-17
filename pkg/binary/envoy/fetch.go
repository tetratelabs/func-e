// Copyright 2019 Tetrate
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package envoy

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"

	"github.com/schollz/progressbar/v3"

	"github.com/tetratelabs/getenvoy/pkg/globals"
	tar "github.com/tetratelabs/getenvoy/pkg/internal"
	"github.com/tetratelabs/getenvoy/pkg/manifest"
	"github.com/tetratelabs/getenvoy/pkg/transport"
)

const binEnvoy = "bin/envoy"

// FetchIfNeeded downloads an Envoy binary corresponding to the given reference and returns a path to it or an error.
func FetchIfNeeded(o *globals.GlobalOpts, reference string) (string, error) {
	ref, err := manifest.ParseReference(reference)
	if err != nil {
		return "", err
	}

	platformPath := filepath.Join(o.HomeDir, "builds", ref.Flavor, ref.Version, ref.Platform)
	envoyPath := filepath.Join(platformPath, binEnvoy)
	_, err = os.Stat(envoyPath)
	switch {
	case os.IsNotExist(err):
		if e := os.MkdirAll(platformPath, 0750); e != nil {
			return "", fmt.Errorf("unable to create directory %q: %w", platformPath, e)
		}

		m, e := manifest.FetchManifest(o.ManifestURL)
		if e != nil {
			return "", e
		}

		downloadLocationURL, e := manifest.LocateBuild(ref, m)
		if e != nil {
			return "", e
		}

		fmt.Fprintln(o.Out, "downloading", downloadLocationURL) //nolint
		if e := untarEnvoy(platformPath, downloadLocationURL, o.Out); e != nil {
			return "", e
		}
	case err == nil:
		fmt.Fprintln(o.Out, ref, "is already downloaded") //nolint
	default:
		// TODO: figure out how to get a stat error that isn't file not exist so we can test this
		return "", err
	}
	return verifyEnvoy(platformPath)
}

func verifyEnvoy(platformPath string) (string, error) {
	envoyPath := filepath.Join(platformPath, binEnvoy)
	stat, err := os.Stat(envoyPath)
	if err != nil {
		return "", err
	}
	if stat.Mode()&0111 == 0 {
		return "", fmt.Errorf("envoy binary not executable at %q", envoyPath)
	}
	return envoyPath, nil
}

func untarEnvoy(dst, url string, out io.Writer) error { // dst, src order like io.Copy
	// #nosec -> url can be anywhere by design
	resp, err := transport.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close() //nolint

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("received %v status code from %s", resp.StatusCode, url)
	}

	// Ensure there's a progress bar while extraction is taking place
	bar := progressbar.NewOptions64(
		resp.ContentLength,
		progressbar.OptionSetWriter(out),
		progressbar.OptionOnCompletion(func() {
			fmt.Fprint(out, "\n")
		}),
	)
	defer bar.Close() //nolint
	src := progressbar.NewReader(resp.Body, bar)

	zSrc, err := tar.NewDecompressor(url, &src)
	if err != nil {
		return err
	}
	if c, ok := zSrc.(io.Closer); ok {
		defer c.Close() //nolint
	}
	return tar.Untar(dst, zSrc)
}
