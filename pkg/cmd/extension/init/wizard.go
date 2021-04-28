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
	"fmt"
	"io"
	"text/template"

	"github.com/manifoldco/promptui"

	uiutil "github.com/tetratelabs/getenvoy/pkg/util/ui"
)

// infoData represents information about user's choice on a single configuration parameter.
type infoData struct {
	Label string
	Value string
}

// defaultInfoStyle returns the default style for rendering user's choices on the screen.
func defaultInfoStyle(styleFuncs template.FuncMap) uiutil.TextStyle {
	return uiutil.Style(styleFuncs, `{{ icon "good" | green }} {{ .Label | italic }} {{ .Value | faint }}`)
}

// wizard implements the interactive mode of the `init` command.
type wizard struct {
	styleFuncs template.FuncMap
	infoStyle  uiutil.TextStyle
	w          io.Writer
}

// newWizard returns a new wizard.
func newWizard(styleFuncs template.FuncMap, w io.Writer) *wizard {
	return &wizard{styleFuncs, defaultInfoStyle(styleFuncs), w}
}

type promptOpts struct {
	init func() error
}

// Fill runs the interactive UI to help a user to fill in parameters.
func (w *wizard) Fill(params *params) error {
	fmt.Fprintln(w.w, uiutil.Underline(w.styleFuncs)("What kind of extension would you like to create?"))
	if err := w.choose(&params.Category, supportedCategories, w.newCategorySelector); err != nil {
		return err
	}
	if err := w.choose(&params.Language, supportedLanguages, w.newLanguageSelector); err != nil {
		return err
	}
	if err := w.prompt(&params.ExtensionDir, w.newExtensionDirPrompt, promptOpts{}); err != nil {
		return err
	}
	if err := w.prompt(&params.Name, w.newNamePrompt, promptOpts{init: params.DefaultName}); err != nil {
		return err
	}
	fmt.Fprintln(w.w, "Great! Let me help you with that!")
	fmt.Fprintln(w.w)
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

func (w *wizard) prompt(param *param, newPrompt func(*param) *promptui.Prompt, opts promptOpts) error {
	// only show the editor when the parameter hasn't been set on the command line or has an invalid value
	if !param.IsValid() {
		if opts.init != nil {
			if err := opts.init(); err != nil {
				return err
			}
		}
		value, err := newPrompt(param).Run()
		if err != nil {
			return err
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

func (w *wizard) newExtensionDirPrompt(param *param) *promptui.Prompt {
	return w.newPrompt("Provide output directory", param)
}

func (w *wizard) newNamePrompt(param *param) *promptui.Prompt {
	return w.newPrompt("Provide extension name", param)
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
	fmt.Fprintln(w.w, w.infoStyle(data))
	return nil
}
