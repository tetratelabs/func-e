// Copyright 2025 Tetrate
// SPDX-License-Identifier: Apache-2.0

package shutdown

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/tetratelabs/func-e/internal/envoy"
)

// EnableHook is an interface for enabling shutdown hooks, with an indicator if admin is required.
type EnableHook func(*envoy.Runtime) error

// DefaultShutdownHooks is a list of shutdown hooks. All hooks, including admin, are registered here.
var DefaultShutdownHooks = []EnableHook{enableNodeCollection, enableAdminDataCollection}

// wrapError wraps an error from using "gopsutil" or returns nil on "not implemented yet".
// We don't err on unimplemented because we don't want to disturb users for unresolvable reasons.
func wrapError(ctx context.Context, err error, field string, pid int32) error {
	if err == nil {
		err = ctx.Err()
	}
	if err != nil && err.Error() != "not implemented yet" { // don't log if it will never work
		return fmt.Errorf("unable to retrieve %s of pid %d: %w", field, pid, err)
	}
	return nil
}

// writeJSON centralizes logic to avoid writing empty files.
func writeJSON(result interface{}, filename string) error {
	sb := new(strings.Builder)
	if err := json.NewEncoder(sb).Encode(result); err != nil {
		return fmt.Errorf("error serializing %v as JSON: %w", sb, err)
	}
	return os.WriteFile(filename, []byte(sb.String()), 0o600)
}
