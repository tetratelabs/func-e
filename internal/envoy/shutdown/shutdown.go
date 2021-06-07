// Copyright 2021 Tetrate
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

package shutdown

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/tetratelabs/getenvoy/internal/envoy"
)

// EnableHooks is a list of functions that enable shutdown hooks
var EnableHooks = []func(*envoy.Runtime) error{enableEnvoyAdminDataCollection, enableNodeCollection}

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
	return os.WriteFile(filename, []byte(sb.String()), 0600)
}
