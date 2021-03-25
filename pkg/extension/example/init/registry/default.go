// Copyright 2020 Tetrate
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

package registry

//go:generate go run github.com/rakyll/statik -p=templates -m -ns=example/init/templates -src=../../../../../data/example/init/templates -a -include=* -f

import (
	"path"

	"github.com/rakyll/statik/fs"

	// force execution of auto generated code
	_ "github.com/tetratelabs/getenvoy/pkg/extension/example/init/registry/templates"
	"github.com/tetratelabs/getenvoy/pkg/extension/workspace/config/extension"
)

func newDefaultRegistry() registry {
	fileSystem, err := fs.NewWithNamespace("example/init/templates")
	if err != nil {
		// must be caught by unit tests
		panic(err)
	}
	return &fsRegistry{
		fs: fileSystem,
		namingScheme: func(category extension.Category, example string) string {
			return "/" + path.Join(category.String(), example)
		},
	}
}
