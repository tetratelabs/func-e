// Copyright 2025 Tetrate
// SPDX-License-Identifier: Apache-2.0

package runtime

import (
	"context"
	"fmt"
	"net/url"
	"os/user"
	"path/filepath"

	"github.com/tetratelabs/func-e/internal/envoy"
	"github.com/tetratelabs/func-e/internal/globals"
	"github.com/tetratelabs/func-e/internal/version"
)

// InitializeGlobalOpts ensures the global options are initialized.
func InitializeGlobalOpts(o *globals.GlobalOpts, envoyVersionsURL, homeDir, platform string) error {
	if o.Platform == "" { // not overridden for tests
		o.Platform = getPlatform(platform)
	}
	var err error
	if o.HomeDir == "" { // not overridden for tests
		if o.HomeDir, err = getHomeDir(homeDir); err != nil { // overridden for tests
			return nil
		}
	}
	if o.EnvoyVersionsURL == "" { // not overridden for tests
		if o.EnvoyVersionsURL, err = getEnvoyVersionsURL(envoyVersionsURL); err != nil {
			return err
		}
	}
	if o.GetEnvoyVersions == nil { // not overridden for tests
		o.GetEnvoyVersions = envoy.NewGetVersions(o.EnvoyVersionsURL, o.Platform, o.Version)
	}
	return nil
}

func getPlatform(platform string) version.Platform {
	if platform != "" { // set by user
		return version.Platform(platform)
	}
	return globals.DefaultPlatform
}

func getHomeDir(homeDir string) (string, error) {
	if homeDir == "" {
		u, err := user.Current()
		if err != nil || u.HomeDir == "" {
			return "", err
		}
		return filepath.Join(u.HomeDir, ".func-e"), nil
	}
	abs, err := filepath.Abs(homeDir)
	if err != nil {
		return "", err
	}
	return abs, nil
}

func getEnvoyVersionsURL(versionsURL string) (string, error) {
	if versionsURL == "" {
		return globals.DefaultEnvoyVersionsURL, nil
	}
	otherURL, err := url.Parse(versionsURL)
	if err != nil || otherURL.Host == "" || otherURL.Scheme == "" {
		return "", fmt.Errorf("%q is not a valid Envoy versions URL", versionsURL)
	}
	return versionsURL, nil
}

// EnsureEnvoyVersion makes sure the Envoy version is set
func EnsureEnvoyVersion(ctx context.Context, o *globals.GlobalOpts) error {
	if o.EnvoyVersion == "" { // not overridden for tests
		return setEnvoyVersion(ctx, o)
	}
	return nil
}
