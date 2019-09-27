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
	"encoding/json"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/tetratelabs/getenvoy/pkg/binary/envoy"
	"github.com/tetratelabs/getenvoy/pkg/binary/envoytest"
)

func TestGetOpenFileStats(t *testing.T) {
	t.Run("creates non-empty files", func(t *testing.T) {
		r, _ := envoy.NewRuntime(EnableOpenFilesDataCollection)
		defer os.RemoveAll(r.DebugStore() + ".tar.gz")
		defer os.RemoveAll(r.DebugStore())
		envoytest.RunKill(r, filepath.Join("testdata", "null.yaml"), time.Second*10)

		files := [...]string{"lsof/lsof.json"}

		for _, file := range files {
			path := filepath.Join(r.DebugStore(), file)
			f, err := os.Stat(path)
			if err != nil {
				t.Errorf("error stating %v: %v", path, err)
			}
			if f.Size() < 1 {
				t.Errorf("file %v was empty", path)
			}
			if strings.HasSuffix(file, ".json") {
				raw, err := ioutil.ReadFile(path)
				if err != nil {
					t.Errorf("error to read the file %v: %v", path, err)
				}
				var is []interface{}
				if err := json.Unmarshal(raw, &is); err != nil {
					t.Errorf("error to unmarshal json string, %v: \"%v\"", err, raw)
				}
				if len(is) < 1 {
					t.Errorf("unmarshaled content is empty, expected to be a non-empty array: \"%v\"", raw)
				}
			}
		}
	})
}
