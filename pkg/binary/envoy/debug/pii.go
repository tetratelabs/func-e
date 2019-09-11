// Copyright 2019 Tetrate
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

package debug

import (
	"crypto/sha256"
	"fmt"

	"bitbucket.org/creachadair/shell"
)

// ProcessLogs process in by filtering it using format and pii, Hash the resulting array of in and return the final array.
// an empty array of strings and an error instance will be returned in the event of an error
func ProcessLogs(logs []string, format string, containsPII map[string]bool, hash func(string) string) ([]string, error) {
	// filter the valid in according to the format str
	fieldNames, ok := shell.Split(format)
	if !ok {
		return []string{}, fmt.Errorf("error in splitting format string: %s", format)
	}

	out := make([]string, 0, len(logs))
	for _, log := range logs {
		fieldValues, ok := shell.Split(log)
		if !ok {
			fmt.Printf("error in splitting log: %s", log)
			continue
		}

		if len(fieldValues) == len(fieldNames) {
			// pick the PII fields and Hash the fields
			for j, name := range fieldNames {
				if containsPII[name] {
					hash := hash(fieldValues[j])
					fieldValues[j] = hash
				}
			}
			out = append(out, shell.Join(fieldValues))
		}
	}
	return out, nil
}

// hash returns the hashed value of s using sha256 hash function
// TODO: salt the Hash
//nolint: unused,deadcode
func hash(s string) string {
	h := sha256.New()
	_, _ = h.Write([]byte(s))
	return string(h.Sum(nil))
}
