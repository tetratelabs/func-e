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

// printer is a contract between the wizard and the command output.
type printer interface {
	Println(...interface{})
}

// infoData represents information about user's choice on a single configuration parameter.
type infoData struct {
	Label string
	Value string
}

// infoStyle is a style for rendering user's choices on the screen.
type infoStyle struct {
	uiutil.TextStyle
}

func (s infoStyle) Format(data *infoData) string {
	return s.TextStyle.Apply(data)
}

// defaultInfoStyle returns the default style for rendering user's choices on the screen.
func defaultInfoStyle() infoStyle {
	return infoStyle{uiutil.Style(fmt.Sprintf(`{{ "%s" | green }} {{ .Label | italic }} {{ .Value | faint }}`, uiutil.IconGood))}
}

// wizard implements the interactive mode of the `init` command.
type wizard struct {
	out       printer
	infoStyle infoStyle
}

// newWizard returns a new wizard.
func newWizard(out printer) *wizard {
	return &wizard{
		out:       out,
		infoStyle: defaultInfoStyle(),
	}
}

// Fill runs the interactive UI to help a user to fill in parameters.
func (w *wizard) Fill(params *params) error {
	w.out.Println(uiutil.Underline("What kind of extension would you like to create?"))
	if err := w.choose(&params.Category, supportedCategories, w.newCategorySelector); err != nil {
		return err
	}
	if err := w.choose(&params.Language, supportedLanguages, w.newLanguageSelector); err != nil {
		return err
	}
	if err := w.prompt(&params.OutputDir, w.newOutputDirPrompt, filepath.Abs); err != nil {
		return err
	}
	w.out.Println("Great! Let me help you with that!")
	w.out.Println()
	return nil
}

func (w *wizard) choose(param *param, options options, newSelect func(*param, options) *promptui.Select) error {
	// only show the editor when the parameter hasn't been set on the command line or has an invalid value
	if !param.IsValid() {
		i, _, err := newSelect(param, options).Run()
		if err != nil {
			return err
		}
		param.Value = options[i].Value
	}
	// always print the effective parameter value
	return w.printChoice(param, options)
}

func (w *wizard) prompt(param *param, newPrompt func(*param) *promptui.Prompt, transform func(string) (string, error)) error {
	// only show the editor when the parameter hasn't been set on the command line or has an invalid value
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
	// always print the effective parameter value
	return w.printInput(param)
}

func (w *wizard) newCategorySelector(param *param, options options) *promptui.Select {
	return w.newSelector("Choose extension category", param, options)
}

func (w *wizard) newLanguageSelector(param *param, options options) *promptui.Select {
	return w.newSelector("Choose programming language", param, options)
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
			Label:    fmt.Sprintf("%s {{ . | bold }}: ", promptui.IconInitial),
			Active:   fmt.Sprintf("%s {{ .DisplayText | underline }}", promptui.IconSelect),
			Inactive: "  {{ .DisplayText }}",
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

func (w *wizard) printChoice(param *param, options options) error {
	return w.print(&infoData{
		Label: param.Title,
		Value: options.ByValue(param.Value).DisplayText,
	})
}

func (w *wizard) printInput(param *param) error {
	return w.print(&infoData{
		Label: param.Title,
		Value: param.Value,
	})
}

func (w *wizard) print(data *infoData) error {
	w.out.Println(w.infoStyle.Format(data))
	return nil
}
