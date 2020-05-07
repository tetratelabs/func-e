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
	"fmt"
)

// icon represents a graphical icon and its plain text substitute.
type icon struct {
	name  string
	rich  string
	plain string
}

var (
	// iconGood indicates a success.
	iconGood = icon{name: "good", rich: "✔", plain: "*"}

	// iconWarn indicates user's attention is necessary.
	iconWarn = icon{name: "warn", rich: "⚠", plain: "!"}

	// iconBad indicates an error.
	iconBad = icon{name: "bad", rich: "✗", plain: "x"}
)

// icons represents a collection of icons.
type icons []icon

func (icons icons) Index() map[string]icon {
	index := map[string]icon{}
	for i := range icons {
		index[icons[i].name] = icons[i]
	}
	return index
}

var (
	supportedIcons = icons{iconGood, iconWarn, iconBad}
)

// iconStyler returns a text/template function that knows how to style
// a given set of icons.
func iconStyler(icons icons) func(string) string {
	index := icons.Index()
	return func(name string) string {
		icon, exists := index[name]
		if !exists {
			return fmt.Sprintf("icon(%s)", name)
		}
		if StylesEnabled {
			return icon.rich
		}
		return icon.plain
	}
}
