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

package flavors

import (
	"bytes"
	"fmt"
	"text/template"
)

// FlavorConfigTemplate - interface to individual flavors.
type FlavorConfigTemplate interface {
	CheckParseParams(params map[string]string) error
	GetTemplate() string
}

// Main repo for templates.
type flavorStore struct {
	// This is a map indexed by flavor pointing to the individual
	// implementaions of each flavor.
	templates map[string]FlavorConfigTemplate
}

var store = flavorStore{templates: make(map[string]FlavorConfigTemplate)}

// AddFlavor - function is used by individual flavors (like postgres)
// to add the flavor to main repo.
func AddFlavor(flavor string, configTemplate FlavorConfigTemplate) {
	store.templates[flavor] = configTemplate
}

// GetFlavor - function returns FlavorConfigTemplate structure associated
// with flavor.
func GetFlavor(flavor string) (FlavorConfigTemplate, error) {
	tmplString, ok := store.templates[flavor]
	if !ok {
		return nil, fmt.Errorf("Cannot find template for flavor %s", flavor)
	}

	return tmplString, nil
}

// CreateConfig - function checks flavor specific parameters, get flavor's template and
// create a config.
func CreateConfig(flavor string, params map[string]string) (string, error) {
	flavorData, err := GetFlavor(flavor)

	if err != nil {
		return "", err
	}

	err = flavorData.CheckParseParams(params)
	if err != nil {
		return "", err
	}

	// NOw run the template substitution
	tmpl := template.New(flavor)
	tmpl, err = tmpl.Parse(flavorData.GetTemplate())
	if err != nil {
		// Template is not supplied by a user, but is compiled-in, so this error should
		// happen only during development time.
		return "", fmt.Errorf("Supplied template for flavor %s is incorrect", flavor)
	}
	var buf bytes.Buffer
	tmpl.Execute(&buf, flavorData) //nolint
	return buf.String(), nil
}
