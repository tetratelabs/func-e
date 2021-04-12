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

import (
	"path"

	exampleTemplates "github.com/tetratelabs/getenvoy/data/example/init"
	"github.com/tetratelabs/getenvoy/pkg/extension/workspace/config/extension"
)

var templatesFs = exampleTemplates.GetTemplates()

func newDefaultRegistry() registry {
	return &fsRegistry{
		fs: templatesFs,
		namingScheme: func(category extension.Category, example string) string {
			return path.Join(category.String(), example)
		},
	}
}
