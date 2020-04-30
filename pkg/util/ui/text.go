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

package ui

import (
	"bytes"
	"fmt"
	"text/template"

	"github.com/manifoldco/promptui"
)

var (
	// StylesEnabled is a knob for turning text styles off in case when
	// runtime environment is not capabale to present them.
	StylesEnabled = true
)

var (
	styleFuncs = makeSubstituteFuncMap(promptui.FuncMap)
)

// TextStyle represents a style that can be applied to data to produce
// rich text representation.
type TextStyle func(data interface{}) string

// Apply returns rich text representation for given data.
func (f TextStyle) Apply(data interface{}) string {
	return f(data)
}

// Style creates a new text style according to a given format string.
func Style(format string) TextStyle {
	tpl, err := template.New("").Funcs(styleFuncs).Parse(format)
	if err != nil {
		// must be caught by unit tests
		panic(err)
	}
	style := &textStyle{template: tpl}
	return style.Apply
}

// textStyle is an implementation of TextStyle on top of text/template.
type textStyle struct {
	template *template.Template
}

func (t *textStyle) Apply(data interface{}) string {
	var buf bytes.Buffer
	err := t.template.Execute(&buf, data)
	if err != nil {
		return fmt.Sprintf("%v", data)
	}
	return buf.String()
}

// makeSubstituteFuncMap returns a substitute func map
// where all coloring functions can be dynamically disabled.
func makeSubstituteFuncMap(source template.FuncMap) template.FuncMap {
	result := template.FuncMap{}
	for name, any := range source {
		if f, ok := any.(func(interface{}) string); ok {
			result[name] = makeSubstituteFunc(f)
		} else {
			result[name] = noStyle
		}
	}
	return result
}

// makeSubstituteFunc returns a substitute func that turns into no op
// when text styles are disabled.
func makeSubstituteFunc(style func(interface{}) string) func(interface{}) string {
	return func(v interface{}) string {
		if StylesEnabled {
			return style(v)
		}
		return noStyle(v)
	}
}

// noStyle applies no style to given data.
func noStyle(v interface{}) string {
	if s, ok := v.(string); ok {
		return s
	}
	return fmt.Sprintf("%v", v)
}
