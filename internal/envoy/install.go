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
	"context"
	"fmt"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"time"

	"github.com/tetratelabs/func-e/internal/globals"
	"github.com/tetratelabs/func-e/internal/moreos"
	"github.com/tetratelabs/func-e/internal/tar"
	"github.com/tetratelabs/func-e/internal/version"
)

var binEnvoy = filepath.Join("bin", "envoy"+moreos.Exe)

// InstallIfNeeded downloads an Envoy binary corresponding to the given version and returns a path to it or an error.
func InstallIfNeeded(ctx context.Context, o *globals.GlobalOpts, v version.Version) (string, error) {
	installPath := filepath.Join(o.HomeDir, "versions", string(v))
	envoyPath := filepath.Join(installPath, binEnvoy)
	_, err := os.Stat(envoyPath)
	switch {
	case os.IsNotExist(err):
		var ev version.ReleaseVersions // Get version metadata for what we will install
		ev, err = FuncEVersions(ctx, o.EnvoyVersionsURL, o.Platform, v)
		if err != nil {
			return "", err
		}

		tarballURL := ev.Versions[v].Tarballs[o.Platform] // Ensure there is a version for this platform
		if tarballURL == "" {
			return "", fmt.Errorf("couldn't find version %q for platform %q", v, o.Platform)
		}

		tarball := version.Tarball(path.Base(string(tarballURL)))
		sha256Sum := ev.SHA256Sums[tarball]
		if len(sha256Sum) != 64 {
			return "", fmt.Errorf("couldn't find sha256Sum of version %q for platform %q: %w", v, o.Platform, err)
		}

		var mtime time.Time // Create a directory for the version, preserving the release date as its mtime
		if mtime, err = time.Parse("2006-01-02", string(ev.Versions[v].ReleaseDate)); err != nil {
			return "", fmt.Errorf("couldn't find releaseDate of version %q for platform %q: %w", v, o.Platform, err)
		}
		if err = os.MkdirAll(installPath, 0750); err != nil {
			return "", fmt.Errorf("unable to create directory %q: %w", installPath, err)
		}

		moreos.Fprintf(o.Out, "downloading %s\n", tarballURL)                                     //nolint
		if err = untarEnvoy(ctx, installPath, tarballURL, sha256Sum, o.Platform, v); err != nil { //nolint
			return "", err
		}
		if err = os.Chtimes(installPath, mtime, mtime); err != nil { // overwrite the mtime to preserve it in the list
			return "", fmt.Errorf("unable to set date of directory %q: %w", installPath, err)
		}
	case err == nil:
		moreos.Fprintf(o.Out, "%s is already downloaded\n", v) //nolint
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
	if !moreos.IsExecutable(stat) {
		return "", fmt.Errorf("envoy binary not executable at %q", envoyPath)
	}
	return envoyPath, nil
}

func untarEnvoy(ctx context.Context, dst string, src version.TarballURL, // dst, src order like io.Copy
	sha256Sum version.SHA256Sum, p version.Platform, v version.Version) error {
	res, err := httpGet(ctx, string(src), p, v)
	if err != nil {
		return err
	}
	defer res.Body.Close() //nolint

	if res.StatusCode != http.StatusOK {
		return fmt.Errorf("received %v status code from %s", res.StatusCode, src)
	}
	if err = tar.UntarAndVerify(dst, res.Body, sha256Sum); err != nil {
		return fmt.Errorf("error untarring %s: %w", src, err)
	}
	return nil
}
