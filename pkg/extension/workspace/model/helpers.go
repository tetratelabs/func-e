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

package model

import (
	"regexp"

	"github.com/pkg/errors"
)

var (
	exampleNameFormat = regexp.MustCompile(`^[a-z0-9._-]+$`)
)

// ValidateExampleName returns an error if a given value is not a valid
// example name.
func ValidateExampleName(text string) error {
	if text == "" {
		return errors.Errorf("example name cannot be empty")
	}
	if !exampleNameFormat.MatchString(text) {
		return errors.Errorf("%q is not a valid example name. Example name must match the format %q."+
			" E.g., 'my.example', 'my-example' or 'my_example'", text, exampleNameFormat)
	}
	return nil
}
