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
	"fmt"
	"path/filepath"

	"github.com/manifoldco/promptui"

	uiutil "github.com/tetratelabs/getenvoy/pkg/util/ui"
)

type printer interface {
	Println(...interface{})
}

type paramStyle struct {
	uiutil.TextStyle
}

func (s paramStyle) Format(param *param) string {
	return s.TextStyle.Format(param)
}

func defaultStyle() paramStyle {
	return paramStyle{uiutil.MustTextStyle(fmt.Sprintf(`{{ "%s" | green }} {{ .Title | italic }} {{ .Value | faint }}`, promptui.IconGood))}
}

type wizard struct {
	out   printer
	style paramStyle
}

func newWizard(out printer) *wizard {
	return &wizard{
		out:   out,
		style: defaultStyle(),
	}
}

func (w *wizard) Fill(params *params) error {
	if err := w.choose(&params.Category, w.newCategorySelector); err != nil {
		return err
	}
	if err := w.choose(&params.Language, w.newLanguageSelector); err != nil {
		return err
	}
	if err := w.prompt(&params.OutputDir, w.newOutputDirPrompt, filepath.Abs); err != nil {
		return err
	}
	return nil
}

func (w *wizard) choose(param *param, newSelect func(*param) *promptui.Select) error {
	if !param.IsValid() {
		_, value, err := newSelect(param).Run()
		if err != nil {
			return err
		}
		param.Value = value
	}
	return w.print(param)
}

func (w *wizard) prompt(param *param, newPrompt func(*param) *promptui.Prompt, transform func(string) (string, error)) error {
	if !param.IsValid() {
		value, err := newPrompt(param).Run()
		if err != nil {
			return err
		}
		if transform != nil {
			value, err = transform(value)
			if err != nil {
				return err
			}
		}
		param.Value = value
	}
	return w.print(param)
}

func (w *wizard) newCategorySelector(param *param) *promptui.Select {
	return w.newSelector("Choose extension category", param, supportedCategories)
}

func (w *wizard) newLanguageSelector(param *param) *promptui.Select {
	return w.newSelector("Choose programming language", param, supportedLanguages)
}

func (w *wizard) newOutputDirPrompt(param *param) *promptui.Prompt {
	return w.newPrompt("Provide output directory", param)
}

func (w *wizard) newSelector(prompt string, param *param, options options) *promptui.Select {
	return &promptui.Select{
		Label:        prompt,
		Items:        options,
		Size:         options.Count(),
		CursorPos:    options.IndexOf(param.Value),
		HideSelected: true,
		Templates: &promptui.SelectTemplates{
			Label: fmt.Sprintf("%s {{ . | bold }}: ", promptui.IconInitial),
		},
	}
}

func (w *wizard) newPrompt(prompt string, param *param) *promptui.Prompt {
	return &promptui.Prompt{
		Label:       prompt,
		Pointer:     promptui.PipeCursor,
		Default:     param.Value,
		AllowEdit:   true,
		Validate:    param.Validator,
		HideEntered: true,
		Templates: &promptui.PromptTemplates{
			Prompt: fmt.Sprintf("%s {{ . | bold }}: ", promptui.IconInitial),
		},
	}
}

func (w *wizard) print(param *param) error {
	w.out.Println(w.style.Format(param))
	return nil
}
