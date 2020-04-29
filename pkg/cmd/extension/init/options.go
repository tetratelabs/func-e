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

// options represents an exhaustive list of valid values.
type options []string

func (o options) Contains(value string) bool {
	return o.IndexOf(value) >= 0
}

func (o options) Count() int {
	return len(o)
}

func (o options) IndexOf(value string) int {
	for i, option := range o {
		if option == value {
			return i
		}
	}
	return -1
}
