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

package cmd_test

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/tetratelabs/func-e/internal/moreos"
)

func (r *runner) Which(ctx context.Context, args []string) error {
	return r.c.RunContext(ctx, args)
}

func TestFuncEWhich(t *testing.T) {
	o := setupTest(t)

	c, stdout, stderr := newApp(o)

	require.NoError(t, c.Run([]string{"func-e", "which"}))
	envoyPath := filepath.Join(o.HomeDir, "versions", o.EnvoyVersion.String(), "bin", "envoy"+moreos.Exe)
	require.Equal(t, moreos.Sprintf("%s\n", envoyPath), stdout.String())
	require.Empty(t, stderr)
}
