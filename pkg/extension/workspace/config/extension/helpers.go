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

package extension

import (
	"encoding/json"

	"github.com/pkg/errors"

	"github.com/tetratelabs/getenvoy/pkg/extension/workspace/config"
	"github.com/tetratelabs/getenvoy/pkg/types"
	"github.com/tetratelabs/multierror"
)

// NewExtensionDescriptor returns a new extension descriptor instance.
func NewExtensionDescriptor() *Descriptor {
	return &Descriptor{
		Meta: config.Meta{
			Kind: Kind,
		},
	}
}

// categories represents a list of extension categories.
type categories []Category

func (categories categories) Index() map[string]Category {
	index := make(map[string]Category, len(categories))
	for _, category := range categories {
		index[category.String()] = category
	}
	return index
}

var (
	// Categories is a list of valid extension categories.
	Categories    = categories{EnvoyHTTPFilter, EnvoyNetworkFilter, EnvoyAccessLogger}
	categoryIndex = Categories.Index()
)

// ParseCategory parses text representation of the extension category.
func ParseCategory(text string) (Category, error) {
	category, valid := categoryIndex[text]
	if !valid {
		return "", errors.Errorf("%q is not a valid extension category", text)
	}
	return category, nil
}

// UnmarshalJSON implements the json.Unmarshaller interface.
func (c *Category) UnmarshalJSON(b []byte) error {
	var value string
	if err := json.Unmarshal(b, &value); err != nil {
		return err
	}
	category, err := ParseCategory(value)
	if err != nil {
		return err
	}
	*c = category
	return nil
}

// languages represents a list of programming languages.
type languages []Language

func (languages languages) Index() map[string]Language {
	index := make(map[string]Language, len(languages))
	for _, language := range languages {
		index[language.String()] = language
	}
	return index
}

var (
	// Languages is a list of valid programming languages.
	Languages     = languages{LanguageRust}
	languageIndex = Languages.Index()
)

// ParseLanguage parses text representation of the extension language.
func ParseLanguage(text string) (Language, error) {
	language, valid := languageIndex[text]
	if !valid {
		return "", errors.Errorf("%q is not a valid programming language", text)
	}
	return language, nil
}

// UnmarshalJSON implements the json.Unmarshaller interface.
func (l *Language) UnmarshalJSON(b []byte) error {
	var value string
	if err := json.Unmarshal(b, &value); err != nil {
		return err
	}
	language, err := ParseLanguage(value)
	if err != nil {
		return err
	}
	*l = language
	return nil
}

// Default sets default values to optional fields.
func (d *Descriptor) Default() {
}

// Validate returns an error if Descriptor is not valid.
func (d *Descriptor) Validate() (errs error) {
	if d.Category == "" {
		errs = multierror.Append(errs, errors.New("extension category cannot be empty"))
	}
	if d.Language == "" {
		errs = multierror.Append(errs, errors.New("programming language cannot be empty"))
	}
	if err := d.Runtime.Validate(); err != nil {
		errs = multierror.Append(errs, errors.Wrap(err, "runtime description is not valid"))
	}
	return
}

// Validate returns an error if Runtime is not valid.
func (r *Runtime) Validate() (errs error) {
	if r.Envoy.Version == "" {
		errs = multierror.Append(errs, errors.New("Envoy version cannot be empty"))
	}
	if r.Envoy.Version != "" {
		if _, err := types.ParseReference(r.Envoy.Version); err != nil {
			errs = multierror.Append(errs, errors.Wrap(err, "Envoy version is not valid"))
		}
	}
	return
}
