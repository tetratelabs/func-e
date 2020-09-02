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

package docker

import "strings"

// ImageName represents a name of a Docker image.
type ImageName struct {
	Org  string
	Name string
	Tag  string
}

func (n ImageName) String() string {
	repository := strings.Join([]string{n.Org, n.Name}, "/")
	if n.Tag == "" {
		return repository
	}
	return strings.Join([]string{repository, n.Tag}, ":")
}
