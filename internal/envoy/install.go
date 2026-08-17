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
	devLatest := v == version.DevLatest
	if devLatest {
		v = version.Dev
	}
	installPath := filepath.Join(o.EnvoyVersionsDir(), v.String())
	envoyPath := filepath.Join(installPath, binEnvoy)
	_, err := os.Stat(envoyPath)

	var evs *version.ReleaseVersions // Get version metadata for what we will install

	if devLatest && err == nil {
		evs, err = o.GetEnvoyVersions(ctx)
		if err != nil {
			return "", err
		}
		if evs.Dev != nil { // Skip re-download if the local install matches the remote release date
			remoteMtime, parseErr := time.Parse("2006-01-02", string(evs.Dev.ReleaseDate))
			stat, statErr := os.Stat(installPath)
			if parseErr == nil && statErr == nil && stat.ModTime().UTC().Truncate(24*time.Hour).Equal(remoteMtime) {
				return verifyEnvoy(installPath)
			}
		}
		err = os.ErrNotExist // force the download branch
	}

	switch {
	case os.IsNotExist(err):
		if evs == nil {
			evs, err = o.GetEnvoyVersions(ctx)
			if err != nil {
				return "", err
			}
		}

		var tarballURL version.TarballURL
		var releaseDate version.ReleaseDate
		if v == version.Dev {
			if evs.Dev != nil {
				tarballURL = evs.Dev.Tarballs[o.Platform]
				releaseDate = evs.Dev.ReleaseDate
			}
		} else {
			r := evs.Versions[v]
			tarballURL = r.Tarballs[o.Platform]
			releaseDate = r.ReleaseDate
		}
		if tarballURL == "" { // Ensure there is a version for this platform
			return "", fmt.Errorf("couldn't find version %q for platform %q", v, o.Platform)
		}

		tarball := version.Tarball(path.Base(string(tarballURL)))
		sha256Sum := evs.SHA256Sums[tarball]
		if len(sha256Sum) != 64 {
			return "", fmt.Errorf("couldn't find sha256Sum of version %q for platform %q: %w", v, o.Platform, err)
		}

		var mtime time.Time // Create a directory for the version, preserving the release date as its mtime
		if mtime, err = time.Parse("2006-01-02", string(releaseDate)); err != nil {
			return "", fmt.Errorf("couldn't find releaseDate of version %q for platform %q: %w", v, o.Platform, err)
		}
		// Unarchive into a staging directory next to the install path, and only move it into place
		// once the SHA-256 sum matches. Otherwise an interrupted download leaves a partial binary
		// at the install path, which later runs treat as already downloaded and never repair.
		versionsDir := o.EnvoyVersionsDir()
		if err = os.MkdirAll(versionsDir, 0o750); err != nil {
			return "", fmt.Errorf("unable to create directory %q: %w", versionsDir, err)
		}
		// The staging directory shares a filesystem with the install path, so the move is atomic,
		// and its name is not a version, so "func-e versions" ignores it if anything is left behind.
		stageDir, err := os.MkdirTemp(versionsDir, ".download-")
		if err != nil {
			return "", fmt.Errorf("unable to create directory in %q: %w", versionsDir, err)
		}
		defer os.RemoveAll(stageDir) //nolint:errcheck // best effort cleanup of a partial download
		stagePath := filepath.Join(stageDir, v.String())

		o.Logf("downloading %s\n", tarballURL)
		if err := untarEnvoy(ctx, o.HTTPClient, stagePath, tarballURL, sha256Sum, o.UserAgent); err != nil {
			return "", err
		}
		if err = os.Chtimes(stagePath, mtime, mtime); err != nil { // overwrite the mtime to preserve it in the list
			return "", fmt.Errorf("unable to set date of directory %q: %w", stagePath, err)
		}
		// A re-download of "dev" has an existing directory in the way of the move.
		if err = os.RemoveAll(installPath); err != nil {
			return "", fmt.Errorf("unable to remove directory %q: %w", installPath, err)
		}
		if err = os.Rename(stagePath, installPath); err != nil {
			return "", fmt.Errorf("unable to move directory %q to %q: %w", stagePath, installPath, err)
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

func untarEnvoy(ctx context.Context, client *http.Client, dst string, src version.TarballURL, // dst, src order like io.Copy
	sha256Sum version.SHA256Sum, ua string,
) error {
	res, err := httpGet(ctx, client, string(src), ua)
	if err != nil {
		return err
	}
	defer res.Body.Close() //nolint:errcheck // body untarred below

	if res.StatusCode != http.StatusOK {
		return fmt.Errorf("received %v status code from %s", res.StatusCode, src)
	}
	if err = tar.UntarAndVerify(dst, res.Body, sha256Sum); err != nil {
		return fmt.Errorf("error untarring %s: %w", src, err)
	}
	return nil
}
