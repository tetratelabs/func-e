// Copyright func-e contributors
// SPDX-License-Identifier: Apache-2.0

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
	"github.com/tetratelabs/func-e/internal/tar"
	"github.com/tetratelabs/func-e/internal/version"
)

var binEnvoy = filepath.Join("bin", "envoy")

// InstallIfNeeded downloads an Envoy binary corresponding to globals.GlobalOpts and returns a path to it or an error.
func InstallIfNeeded(ctx context.Context, o *globals.GlobalOpts) (string, error) {
	v := o.EnvoyVersion
	installPath := filepath.Join(o.EnvoyVersionsDir(), v.String())
	envoyPath := filepath.Join(installPath, binEnvoy)
	_, err := os.Stat(envoyPath)
	switch {
	case os.IsNotExist(err):
		var evs *version.ReleaseVersions // Get version metadata for what we will install
		evs, err = o.GetEnvoyVersions(ctx)
		if err != nil {
			return "", err
		}

		tarballURL := evs.Versions[v].Tarballs[o.Platform] // Ensure there is a version for this platform
		if tarballURL == "" {
			return "", fmt.Errorf("couldn't find version %q for platform %q", v, o.Platform)
		}

		tarball := version.Tarball(path.Base(string(tarballURL)))
		sha256Sum := evs.SHA256Sums[tarball]
		if len(sha256Sum) != 64 {
			return "", fmt.Errorf("couldn't find sha256Sum of version %q for platform %q: %w", v, o.Platform, err)
		}

		var mtime time.Time // Create a directory for the version, preserving the release date as its mtime
		if mtime, err = time.Parse("2006-01-02", string(evs.Versions[v].ReleaseDate)); err != nil {
			return "", fmt.Errorf("couldn't find releaseDate of version %q for platform %q: %w", v, o.Platform, err)
		}
		if err = os.MkdirAll(installPath, 0o750); err != nil {
			return "", fmt.Errorf("unable to create directory %q: %w", installPath, err)
		}
		o.Logf("downloading %s\n", tarballURL)
		if err = untarEnvoy(ctx, installPath, tarballURL, sha256Sum, o.Platform, o.Version); err != nil { //nolint
			return "", err
		}
		if err = os.Chtimes(installPath, mtime, mtime); err != nil { // overwrite the mtime to preserve it in the list
			return "", fmt.Errorf("unable to set date of directory %q: %w", installPath, err)
		}
	case err == nil:
		o.Logf("%s is already downloaded\n", v)
	default:
		// TODO: figure out how to Get a stat error that isn't file not exist so we can test this
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
	if stat.Mode()&0o111 == 0 { // isExecutable
		return "", fmt.Errorf("envoy binary not executable at %q", envoyPath)
	}
	return envoyPath, nil
}

func untarEnvoy(ctx context.Context, dst string, src version.TarballURL, // dst, src order like io.Copy
	sha256Sum version.SHA256Sum, p version.Platform, v string,
) error {
	res, err := httpGet(ctx, http.DefaultClient, string(src), p, v)
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
