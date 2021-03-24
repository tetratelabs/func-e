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

import (
	"net/http"
	"path"
	"sync"

	extensionTemplates "github.com/tetratelabs/getenvoy/data/extension/init"
	"github.com/tetratelabs/getenvoy/pkg/extension/workspace/config/extension"
)

var (
	// source of extension templates
	templates     templateSource
	templatesOnce sync.Once
)

var templatesFs = extensionTemplates.GetTemplates()

// getTemplateSource returns a source of extension templates.
func getTemplateSource() templateSource {
	// do lazy init to avoid zip decompression unless absolutely necessary
	templatesOnce.Do(func() {
		templates = &fsTemplateSource{
			fs: http.FS(templatesFs),
			namingScheme: func(language extension.Language, category extension.Category, template string) string {
				return "/" + path.Join(language.String(), category.String(), template)
			},
		}
	})
	return templates
}
