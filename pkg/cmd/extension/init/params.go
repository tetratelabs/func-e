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
	"errors"
	"fmt"
	"os"
	"path/filepath"

	scaffold "github.com/tetratelabs/getenvoy/pkg/extension/init"
	"github.com/tetratelabs/getenvoy/pkg/extension/workspace/config/extension"
	osutil "github.com/tetratelabs/getenvoy/pkg/util/os"
)

// TODO see if all these parameter fields need to be upper-cased
// param represents a single parameter to be filled in by a user.
type param struct {
	Title     string
	Value     string
	Validator func(string) error
}

func (p *param) Validate() error {
	return p.Validator(p.Value)
}

func (p *param) IsValid() bool {
	return p.Validate() == nil
}

// params represents all parameters to be filled in by a user.
type params struct {
	Category     param
	Language     param
	ExtensionDir param
	Name         param
}

func (o *params) Validate() error {
	if err := o.Category.Validate(); err != nil {
		return err
	}
	if err := o.Language.Validate(); err != nil {
		return err
	}
	if err := o.ExtensionDir.Validate(); err != nil {
		return err
	}
	if err := o.Name.Validate(); err != nil {
		return err
	}
	return nil
}

func (o *params) DefaultName() error {
	if o.Name.Value != "" {
		return errors.New("empty name")
	}
	category, err := extension.ParseCategory(o.Category.Value)
	if err != nil {
		return err
	}
	extensionDir, err := filepath.Abs(o.ExtensionDir.Value)
	if err != nil {
		return err
	}
	o.Name.Value = scaffold.GenerateExtensionName(category, extensionDir)
	return nil
}

//nolint:gocyclo
func newParams() *params {
	return &params{
		Category: param{
			Title: "Category",
			Validator: func(value string) error {
				if value == "" {
					return errors.New("extension category cannot be empty")
				}
				if !supportedCategories.Contains(value) {
					return fmt.Errorf("%q is not a supported extension category", value)
				}
				return nil
			},
		},
		Language: param{
			Title: "Language",
			Validator: func(value string) error {
				if value == "" {
					return errors.New("programming language cannot be empty")
				}
				if !supportedLanguages.Contains(value) {
					return fmt.Errorf("%q is not a supported programming language", value)
				}
				return nil
			},
		},
		ExtensionDir: param{
			Title: "Extension directory",
			Validator: func(value string) error {
				extensionDir, err := filepath.Abs(value)
				if err != nil {
					return err
				}
				info, err := os.Stat(extensionDir)
				if err != nil {
					if os.IsNotExist(err) {
						return nil
					}
					return err
				}
				if !info.IsDir() {
					return fmt.Errorf("extension directory is a file: %s", extensionDir)
				}
				empty, err := osutil.IsEmptyDir(extensionDir)
				if err != nil {
					return err
				}
				if !empty {
					return fmt.Errorf("extension directory must be empty or new: %s", extensionDir)
				}
				return nil
			},
		},
		Name: param{
			Title: "Extension name",
			Validator: func(value string) error {
				if value == "" {
					return errors.New("extension name cannot be empty")
				}
				return extension.ValidateExtensionName(value)
			},
		},
	}
}
