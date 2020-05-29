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
	"os"
	"path/filepath"

	"github.com/pkg/errors"

	osutil "github.com/tetratelabs/getenvoy/pkg/util/os"
)

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
	Category  param
	Language  param
	OutputDir param
}

func (o *params) Validate() error {
	if err := o.Category.Validate(); err != nil {
		return err
	}
	if err := o.Language.Validate(); err != nil {
		return err
	}
	if err := o.OutputDir.Validate(); err != nil {
		return err
	}
	return nil
}

func newParams() *params {
	return &params{
		Category: param{
			Title: "Category",
			Validator: func(value string) error {
				if !supportedCategories.Contains(value) {
					return errors.Errorf("%q is not a supported extension category", value)
				}
				return nil
			},
		},
		Language: param{
			Title: "Language",
			Validator: func(value string) error {
				if !supportedLanguages.Contains(value) {
					return errors.Errorf("%q is not a supported programming language", value)
				}
				return nil
			},
		},
		OutputDir: param{
			Title: "Output directory",
			Validator: func(value string) error {
				value = filepath.Clean(value)
				outputDir, err := filepath.Abs(value)
				if err != nil {
					return err
				}
				info, err := os.Stat(outputDir)
				if err != nil {
					if os.IsNotExist(err) {
						return nil
					}
					return err
				}
				if !info.IsDir() {
					return errors.Errorf("output path is not a directory: %s", outputDir)
				}
				empty, err := osutil.IsEmptyDir(outputDir)
				if err != nil {
					return err
				}
				if !empty {
					return errors.Errorf("output directory must be empty or new: %s", outputDir)
				}
				return nil
			},
		},
	}
}
