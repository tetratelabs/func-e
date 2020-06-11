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
	"fmt"
	"net/http"

	"github.com/hashicorp/go-multierror"

	"github.com/tetratelabs/getenvoy/pkg/extension/workspace/config/extension"
)

// templateSource represents a source of extension templates.
type templateSource interface {
	// GetTemplateDir returns a directory with template source code.
	GetTemplateDir(language extension.Language, category extension.Category, template string) (string, http.FileSystem, error)
}

// fsTemplateSource represents a source of extension templates backend by
// an in-memory file system.
type fsTemplateSource struct {
	fs           http.FileSystem
	namingScheme func(language extension.Language, category extension.Category, template string) string
}

func (s *fsTemplateSource) GetTemplateDir(language extension.Language, category extension.Category,
	template string) (dirName string, fs http.FileSystem, errs error) {
	dirName = s.namingScheme(language, category, template)
	dir, err := s.fs.Open(dirName)
	if err != nil {
		return "", nil, err
	}
	defer func() {
		if e := dir.Close(); e != nil {
			errs = multierror.Append(errs, e)
		}
	}()
	info, err := dir.Stat()
	if err != nil {
		return "", nil, err
	}
	if !info.IsDir() {
		return "", nil, fmt.Errorf("%q is not a directory", dirName)
	}
	return dirName, s.fs, nil
}
