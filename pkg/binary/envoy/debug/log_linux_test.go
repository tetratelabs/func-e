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
	"os"
	"path/filepath"
	"testing"

	"github.com/tetratelabs/getenvoy/pkg/binary/envoy"
	"github.com/tetratelabs/getenvoy/pkg/binary/envoytest"
	"github.com/tetratelabs/getenvoy/pkg/manifest"
)

func Test_capture(t *testing.T) {
	t.Run("creates non-empty access and error log files", func(t *testing.T) {
		key, _ := manifest.NewKey(envoytest.Reference)
		// EnableEnvoyAdminDataCollection is here to create some requests to the admin endpoint for stdout
		r, _ := envoy.NewRuntime(EnableEnvoyLogCollection, EnableEnvoyAdminDataCollection)
		defer os.RemoveAll(r.DebugStore() + ".tar.gz")
		defer os.RemoveAll(r.DebugStore())
		startWaitKillUnarchiveGetEnvoy(r, key, filepath.Join("testdata", "stdout.yaml"))

		for _, filename := range []string{"logs/access.log", "logs/error.log"} {
			path := filepath.Join(r.DebugStore(), filename)
			f, err := os.Stat(path)
			if err != nil {
				t.Errorf("error stating %v: %v", path, err)
			}
			if f.Size() < 1 {
				t.Errorf("file %v was empty", path)
			}
		}
	})
}
