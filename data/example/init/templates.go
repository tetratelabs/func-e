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

package init

import (
	"embed"
	"io/fs"
)

// templatesFs includes only the relative path of "templates".
//
// See RATIONALE.md for more information on embedding
//go:embed templates/*
var templatesFs embed.FS

// GetTemplates returns the templates directory as a filesystem
func GetTemplates() fs.FS {
	f, err := fs.Sub(templatesFs, "templates")
	if err != nil {
		panic(err) // unexpected or a typo
	}
	return f
}
