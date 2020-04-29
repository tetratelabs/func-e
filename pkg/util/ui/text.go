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

// TextStyle represents text formatting rules.
type TextStyle interface {
	Format(data interface{}) string
}

// MustTextStyle creates a new text style according to a given format string.
func MustTextStyle(style string) TextStyle {
	tpl, err := template.New("").Funcs(promptui.FuncMap).Parse(style)
	if err != nil {
		panic(err)
	}
	return &textStyle{template: tpl}
}

type textStyle struct {
	template *template.Template
}

func (t *textStyle) Format(data interface{}) string {
	var buf bytes.Buffer
	err := t.template.Execute(&buf, data)
	if err != nil {
		return fmt.Sprintf("%v", data)
	}
	return buf.String()
}
