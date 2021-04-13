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

package config

import (
	"sigs.k8s.io/yaml"
)

var (
	// Marshal converts an object into its YAML representation.
	Marshal = yaml.Marshal
	// Unmarshal converts YAML into an object.
	Unmarshal = func(bytes []byte, i interface{}) error {
		return yaml.Unmarshal(bytes, i) // no opts
	}
)
