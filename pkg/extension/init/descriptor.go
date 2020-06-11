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
	"github.com/tetratelabs/getenvoy/pkg/extension/workspace/config/extension"
)

// NewExtension returns a new extension descriptor.
func NewExtension(category, language string) (*extension.Descriptor, error) {
	extensionCategory, err := extension.ParseCategory(category)
	if err != nil {
		return nil, err
	}
	extensionLanguage, err := extension.ParseLanguage(language)
	if err != nil {
		return nil, err
	}
	return &extension.Descriptor{
		Category: extensionCategory,
		Language: extensionLanguage,
	}, nil
}
