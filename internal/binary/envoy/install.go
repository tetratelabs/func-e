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

	"github.com/tetratelabs/getenvoy/internal/errors"
	"github.com/tetratelabs/getenvoy/internal/globals"
	"github.com/tetratelabs/getenvoy/internal/manifest"
	"github.com/tetratelabs/getenvoy/internal/tar"
	"github.com/tetratelabs/getenvoy/internal/transport"
)

const binEnvoy = "bin/envoy"

// InstallIfNeeded downloads an Envoy binary corresponding to the given version and returns a path to it or an error.
func InstallIfNeeded(o *globals.GlobalOpts, goos, version string) (string, error) {
	installPath := filepath.Join(o.HomeDir, "versions", version)
	envoyPath := filepath.Join(installPath, binEnvoy)
	_, err := os.Stat(envoyPath)
	switch {
	case os.IsNotExist(err):
		if e := os.MkdirAll(installPath, 0750); e != nil {
			return "", fmt.Errorf("unable to create directory %q: %w", installPath, e)
		}

		m, e := manifest.GetManifest(o.ManifestURL)
		if e != nil {
			return "", e
		}

		downloadLocationURL, e := downloadURL(m, goos, version)
		if e != nil {
			return "", e
		}

		fmt.Fprintln(o.Out, "downloading", downloadLocationURL) //nolint
		if e := untarEnvoy(installPath, downloadLocationURL, o.Out); e != nil {
			return "", e
		}
	case err == nil:
		fmt.Fprintln(o.Out, version, "is already downloaded") //nolint
	default:
		// TODO: figure out how to get a stat error that isn't file not exist so we can test this
		return "", err
	}
	return verifyEnvoy(installPath)
}

// LocateBuild returns the downloadLocationURL of the associated envoy binary in the manifest using the input key
func downloadURL(m *manifest.Manifest, goos, version string) (string, error) {
	errorNoVersions := errors.NewValidationError("couldn't find version %q for platform %q", version, goos)
	platform := manifest.BuildPlatform(goos)
	if platform == "" {
		return "", errorNoVersions
	}

	// Error if the only released "flavor" or it has no version
	if m.Flavors["standard"] == nil || len(m.Flavors["standard"].Versions) == 0 {
		return "", errorNoVersions
	}

	// Error if the version doesn't exist or has no builds for this platform
	builds := m.Flavors["standard"].Versions[version]
	if builds == nil || builds.Builds == nil || builds.Builds[platform] == nil {
		return "", errorNoVersions
	}

	return builds.Builds[platform].DownloadLocationURL, nil
}

func verifyEnvoy(installPath string) (string, error) {
	envoyPath := filepath.Join(installPath, binEnvoy)
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
	resp, e := transport.Get(url)
	if e != nil {
		return e
	}
	defer resp.Body.Close() //nolint

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("received %v status code from %s", resp.StatusCode, url)
	}

	// Ensure there's a progress while extraction is taking place
	src := progressReader(out, resp.Body, resp.ContentLength)
	defer src.Close() //nolint

	if e = tar.Untar(dst, src); e != nil {
		return fmt.Errorf("error untarring %s: %w", url, e)
	}
	return nil
}

// progressReader will show a spinner or progress bar, depending on if max == -1
func progressReader(dst io.Writer, src io.Reader, max int64) io.ReadCloser {
	b := progressbar.NewOptions64(max,
		progressbar.OptionSetWriter(dst),
		progressbar.OptionOnCompletion(func() {
			fmt.Fprint(dst, "\n")
		}),
	)
	br := progressbar.NewReader(src, b)
	return &br
}
