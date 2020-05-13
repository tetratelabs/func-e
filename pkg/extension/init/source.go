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

package init

//go:generate go run github.com/rakyll/statik -p=templates -m -ns=extension/init/templates -src=../../../data/extension/init/templates -a -include=*

import (
	"path"
	"sync"

	"github.com/rakyll/statik/fs"

	// force execution of auto generated code
	_ "github.com/tetratelabs/getenvoy/pkg/extension/init/templates"
)

var (
	// source of extension templates
	templates     templateSource
	templatesOnce sync.Once
)

// getTemplateSource returns a source of extension templates.
func getTemplateSource() templateSource {
	// do lazy init to avoid zip decompression unless absolutely necessary
	templatesOnce.Do(func() {
		fileSystem, err := fs.NewWithNamespace("extension/init/templates")
		if err != nil {
			panic(err)
		}
		templates = &fsTemplateSource{
			fs: fileSystem,
			namingScheme: func(language, category, template string) string {
				return "/" + path.Join(language, category, template)
			},
		}
	})
	return templates
}
