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

package debug

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/tetratelabs/getenvoy/pkg/binary/envoy"
	"github.com/tetratelabs/getenvoy/pkg/binary/envoytest"
)

func TestMain(m *testing.M) {
	if err := envoytest.Fetch(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	os.Exit(m.Run())
}

// This test relies on a local Envoy binary, if not present it will fetch one from GetEnvoy
// This is more of an integration test than a unit test, but either way is necessary.
func Test_retrieveAdminAPIData(t *testing.T) {
	t.Run("creates all non-empty files", func(t *testing.T) {
		r, err := envoy.NewRuntime(EnableEnvoyAdminDataCollection)
		defer os.RemoveAll(r.DebugStore())
		require.NoError(t, err, "error getting envoy runtime")

		err = envoytest.RunKill(r, filepath.Join("testdata", "null.yaml"), time.Second*10)
		require.NoError(t, err, "error from envoytest.RunKill")

		for _, filename := range adminAPIPaths {
			path := filepath.Join(r.DebugStore(), filename)
			f, err := os.Stat(path)
			require.NoError(t, err, "error stating %v", path)
			require.NotEmpty(t, f.Size(), "file %v was empty", path)
		}
	})

}
