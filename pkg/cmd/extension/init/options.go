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

// option represents a valid configuration value.
type option struct {
	Value       string
	DisplayText string
}

// options represents an exhaustive list of valid configuration values.
type options []option

func (o options) Contains(value string) bool {
	return o.IndexOf(value) >= 0
}

func (o options) ByValue(value string) *option {
	if i := o.IndexOf(value); i >= 0 {
		return &o[i]
	}
	return nil
}

func (o options) IndexOf(value string) int {
	for i := range o {
		if o[i].Value == value {
			return i
		}
	}
	return -1
}

func (o options) Values() []string {
	values := make([]string, len(o))
	for i := range o {
		values[i] = o[i].Value
	}
	return values
}

func (o options) Count() int {
	return len(o)
}
