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
	"strings"

	"github.com/spf13/cobra"
)

// NewCmd returns a command that aggregates all extension-related commands.
func NewCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "extension",
		Short: "Delve into Envoy extensions.",
		Long:  `Explore ready-to-use Envoy extensions or develop a new one.`,
	}
	cmd.AddCommand(NewInitCmd())
	return cmd
}

func hintOneOf(values ...string) string {
	texts := make([]string, len(values))
	for i := range values {
		texts[i] = fmt.Sprintf("%q", values[i])
	}
	return "One of: " + strings.Join(texts, ", ")
}
