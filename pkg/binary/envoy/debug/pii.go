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

var istioFormat = `[%START_TIME%] "%REQ(:METHOD)% %REQ(X-ENVOY-ORIGINAL-PATH?:PATH)% %PROTOCOL%"` +
	` %RESPONSE_CODE% %RESPONSE_FLAGS% "%DYNAMIC_METADATA(istio.mixer:status)%" "%REQ(USER-AGENT)%"`
var defaultPII = map[string]bool{
	"[%START_TIME%]": true,
}

// Filter filters log fields using pii and modify all PII fields using f
type Filter struct {
	f      func(string) string
	pii    map[string]bool
	format []string
}

// NewFilter constructs a custom filter object
func NewFilter(formatStr string, hash func(string) string, pii map[string]bool) (Filter, error) {
	// splitting formats and handle error
	format, ok := shell.Split(formatStr)
	if !ok {
		return Filter{}, fmt.Errorf("error in splitting format string: %s", format)
	}
	return Filter{f: hash, pii: pii, format: format}, nil
}

// Process logs and hash the filter, assumes that filter has valid fields
func (f Filter) Process(logs []string) []string {
	out := make([]string, 0, len(logs))
	for _, log := range logs {
		fieldValues, ok := shell.Split(log)
		if !ok {
			fmt.Printf("error in splitting log: %s", log)
			continue
		}

		if len(fieldValues) == len(f.format) {
			// pick the PII fields and Hash the fields
			for j, name := range f.format {
				// TODO: may require further parsing name to verify if it matches as a key in the map with variables
				if f.pii[name] {
					fieldValues[j] = f.f(fieldValues[j])
				}
			}
			out = append(out, shell.Join(fieldValues))
		}
	}
	return out
}

// ProcessLogs process logs with the default filter an empty array of logs
// and an error instance is returned in an event of error
func ProcessLogs(logs []string) ([]string, error) {
	filter, err := NewFilter(istioFormat, defaultHash, defaultPII)

	if err != nil {
		return []string{}, err
	}
	return filter.Process(logs), nil
}

// defaultHash returns the hashed value of s using sha256 defaultHash function
// TODO: salt the Hash
//nolint: unused,deadcode
func defaultHash(s string) string {
	h := sha256.New()
	_, _ = h.Write([]byte(s))
	return string(h.Sum(nil))
}
