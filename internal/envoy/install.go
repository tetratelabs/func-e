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
	"time"

	"github.com/schollz/progressbar/v3"

	"github.com/tetratelabs/getenvoy/internal/globals"
	"github.com/tetratelabs/getenvoy/internal/tar"
	"github.com/tetratelabs/getenvoy/internal/version"
)

const binEnvoy = "bin/envoy"

// InstallIfNeeded downloads an Envoy binary corresponding to the given version and returns a path to it or an error.
func InstallIfNeeded(o *globals.GlobalOpts, p, v string) (string, error) {
	installPath := filepath.Join(o.HomeDir, "versions", v)
	envoyPath := filepath.Join(installPath, binEnvoy)
	_, err := os.Stat(envoyPath)
	switch {
	case os.IsNotExist(err):
		var ev version.EnvoyVersions // Get version metadata for what we will install
		ev, err = GetEnvoyVersions(o.EnvoyVersionsURL, o.UserAgent)
		if err != nil {
			return "", err
		}

		tarballURL := ev.Versions[v].Tarballs[p] // Ensure there is a version for this platform
		if tarballURL == "" {
			return "", fmt.Errorf("couldn't find version %q for platform %q", v, p)
		}

		var mtime time.Time // Create a directory for the version, preserving the release date as its mtime
		if mtime, err = time.Parse("2006-01-02", ev.Versions[v].ReleaseDate); err != nil {
			return "", fmt.Errorf("couldn't find releaseDate of version %q for platform %q: %w", v, p, err)
		}
		if err = os.MkdirAll(installPath, 0750); err != nil {
			return "", fmt.Errorf("unable to create directory %q: %w", installPath, err)
		}
		if err = os.Chtimes(installPath, mtime, mtime); err != nil {
			return "", fmt.Errorf("unable to set date of directory %q: %w", installPath, err)
		}

		fmt.Fprintln(o.Out, "downloading", tarballURL) //nolint
		if err = untarEnvoy(installPath, tarballURL, o.UserAgent, o.Out); err != nil {
			return "", err
		}
	case err == nil:
		fmt.Fprintln(o.Out, v, "is already downloaded") //nolint
	default:
		// TODO: figure out how to get a stat error that isn't file not exist so we can test this
		return "", err
	}
	return verifyEnvoy(installPath)
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

func untarEnvoy(dst, url, userAgent string, out io.Writer) error { // dst, src order like io.Copy
	resp, err := httpGet(url, userAgent)
	if err != nil {
		return err
	}
	defer resp.Body.Close() //nolint

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("received %v status code from %s", resp.StatusCode, url)
	}

	// Ensure there's a progress while extraction is taking place
	src := progressReader(out, resp.Body, resp.ContentLength)
	defer src.Close() //nolint

	if err = tar.Untar(dst, src); err != nil {
		return fmt.Errorf("error untarring %s: %w", url, err)
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
