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

package example

import (
	"fmt"
	"io"

	"github.com/spf13/cobra"

	scaffoldutil "github.com/tetratelabs/getenvoy/pkg/util/scaffold"
	uiutil "github.com/tetratelabs/getenvoy/pkg/util/ui"
)

// NewAddExampleFeedback returns a progress sink for "add example" operation.
func NewAddExampleFeedback(cmd *cobra.Command) scaffoldutil.ProgressSink {
	return &addExampleFeedback{cmd.ErrOrStderr()}
}

// addExampleFeedback communicates to a user progress of the "add example" operation.
type addExampleFeedback struct {
	w io.Writer
}

func (f *addExampleFeedback) OnStart() {
	fmt.Fprintln(f.w, uiutil.Underline("Scaffolding a new example setup:"))
}

func (f *addExampleFeedback) OnFile(file string) {
	fmt.Fprintln(f.w, uiutil.Style(`{{ icon "good" | green }} {{ . }}`).Apply(file))
}

func (f *addExampleFeedback) OnComplete() {
	fmt.Fprintln(f.w, "Done!")
}

// NewRemoveExampleFeedback returns a progress sink for "remove example" operation.
func NewRemoveExampleFeedback(cmd *cobra.Command) scaffoldutil.ProgressSink {
	return &removeExampleFeedback{cmd.ErrOrStderr()}
}

// removeExampleFeedback communicates to a user progress of the "remove example" operation.
type removeExampleFeedback struct {
	w io.Writer
}

func (f *removeExampleFeedback) OnStart() {
	fmt.Fprintln(f.w, uiutil.Underline("Removing example setup:"))
}

func (f *removeExampleFeedback) OnFile(file string) {
	fmt.Fprintln(f.w, uiutil.Style(`{{ icon "good" | green }} {{ . }}`).Apply(file))
}

func (f *removeExampleFeedback) OnComplete() {
	fmt.Fprintln(f.w, "Done!")
}
