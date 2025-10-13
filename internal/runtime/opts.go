// Copyright 2025 Tetrate
// SPDX-License-Identifier: Apache-2.0

package runtime

import (
	"context"
	"fmt"
	"net/url"
	"os/user"
	"path/filepath"
	"strings"
	"time"

	"github.com/tetratelabs/func-e/internal/envoy"
	"github.com/tetratelabs/func-e/internal/globals"
	"github.com/tetratelabs/func-e/internal/version"
)

// InitializeGlobalOpts ensures the global options are initialized.
func InitializeGlobalOpts(o *globals.GlobalOpts, envoyVersionsURL, homeDir, configHome, dataHome, stateHome, runtimeDir, platform, runID string) error {
	if o.Platform == "" { // not overridden for tests
		o.Platform = getPlatform(platform)
	}

	var err error

	// Legacy mode: FUNC_E_HOME sets all four directories to the same value
	if homeDir != "" {
		abs, err := filepath.Abs(homeDir)
		if err != nil {
			return err
		}
		o.HomeDir = homeDir
		o.ConfigHome = abs
		o.DataHome = abs
		o.StateHome = abs
		o.RuntimeDir = abs
	} else {
		// Independent directories with proper defaults
		if o.ConfigHome == "" { // not overridden for tests
			if o.ConfigHome, err = getConfigHome(configHome); err != nil {
				return err
			}
		}
		if o.DataHome == "" { // not overridden for tests
			if o.DataHome, err = getDataHome(dataHome); err != nil {
				return err
			}
		}
		if o.StateHome == "" { // not overridden for tests
			if o.StateHome, err = getStateHome(stateHome); err != nil {
				return err
			}
		}
		if o.RuntimeDir == "" { // not overridden for tests
			if o.RuntimeDir, err = getRuntimeDir(runtimeDir); err != nil {
				return err
			}
		}
	}

	// Generate or validate runID
	if runID == "" {
		o.RunID = o.GenerateRunID(time.Now())
	} else {
		// Validate that runID doesn't contain path separators
		if strings.ContainsAny(runID, "/\\") {
			return fmt.Errorf("runID cannot contain path separators (/ or \\): %q", runID)
		}
		o.RunID = runID
	}

	if o.EnvoyVersionsURL == "" { // not overridden for tests
		if o.EnvoyVersionsURL, err = getEnvoyVersionsURL(envoyVersionsURL); err != nil {
			return err
		}
	}
	if o.GetEnvoyVersions == nil { // not overridden for tests
		o.GetEnvoyVersions = envoy.NewGetVersions(o.EnvoyVersionsURL, o.Platform, o.Version)
	}

	// Create base XDG directories now that all paths are configured
	return o.Mkdirs()
}

func getPlatform(platform string) version.Platform {
	if platform != "" { // set by user
		return version.Platform(platform)
	}
	return globals.DefaultPlatform
}

func getConfigHome(configHome string) (string, error) {
	if configHome == "" {
		u, err := user.Current()
		if err != nil || u.HomeDir == "" {
			return "", err
		}
		return filepath.Join(u.HomeDir, ".config", "func-e"), nil
	}
	abs, err := filepath.Abs(configHome)
	if err != nil {
		return "", err
	}
	return abs, nil
}

func getDataHome(dataHome string) (string, error) {
	if dataHome == "" {
		u, err := user.Current()
		if err != nil || u.HomeDir == "" {
			return "", err
		}
		return filepath.Join(u.HomeDir, ".local", "share", "func-e"), nil
	}
	abs, err := filepath.Abs(dataHome)
	if err != nil {
		return "", err
	}
	return abs, nil
}

func getStateHome(stateHome string) (string, error) {
	if stateHome == "" {
		u, err := user.Current()
		if err != nil || u.HomeDir == "" {
			return "", err
		}
		return filepath.Join(u.HomeDir, ".local", "state", "func-e"), nil
	}
	abs, err := filepath.Abs(stateHome)
	if err != nil {
		return "", err
	}
	return abs, nil
}

func getRuntimeDir(runtimeDir string) (string, error) {
	if runtimeDir == "" {
		u, err := user.Current()
		if err != nil || u.Uid == "" {
			return "", err
		}
		return filepath.Join("/tmp", fmt.Sprintf("func-e-%s", u.Uid)), nil
	}
	abs, err := filepath.Abs(runtimeDir)
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
