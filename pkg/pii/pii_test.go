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

package pii

import (
	"reflect"
	"strings"
	"testing"
)

func TestProcess(t *testing.T) {
	format := `[%START_TIME%] "%REQ(:METHOD)% %REQ(X-ENVOY-ORIGINAL-PATH?:PATH)% %PROTOCOL%"` +
		` %RESPONSE_CODE% %RESPONSE_FLAGS% "%DYNAMIC_METADATA(istio.mixer:status)%" "%REQ(USER-AGENT)%"`

	validLog := `[2019-09-05T17:53:36.908Z] "GET /productpage HTTP/1.1" 200 - "-"` +
		` "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/76.0.3809.132 Safari/537.36"`
	logWithOneField := `[2019-09-05T17:53:36.908Z]`

	// test cases wrapped up in a map where key is the test case name and value is the struct
	// defining the inputs and expecte outcome
	tests := map[string]struct {
		// input fields
		in     []string
		format string
		pii    map[string]bool

		// output fields
		outLogs []string
		outErr  string
	}{
		// positive test cases
		// only return an array containing one valid log with [%START_TIME%] field hashed
		"simple": {
			in:     []string{validLog, logWithOneField},
			format: format,
			pii:    map[string]bool{"[%START_TIME%]": true},
			// expected expression in "" replaced by ''
			outLogs: []string{`pii 'GET /productpage HTTP/1.1' 200 - - 'Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko)` +
				` Chrome/76.0.3809.132 Safari/537.36'`},
			outErr: "",
		},
		"multiple_pii_fields": {
			in:     []string{validLog, logWithOneField},
			format: format,
			pii: map[string]bool{
				"[%START_TIME%]": true,
				"%REQ(:METHOD)% %REQ(X-ENVOY-ORIGINAL-PATH?:PATH)% %PROTOCOL%": true,
				"%RESPONSE_CODE%": true,
			},
			outLogs: []string{`pii pii pii - - 'Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/76.0.3809.132 Safari/537.36'`},
			outErr:  "",
		},

		// negative test cases
		"empty_logs": {
			in:      []string{},
			format:  format,
			pii:     map[string]bool{"[%START_TIME%]": true},
			outLogs: []string{},
			outErr:  "",
		},
		"all_invalid_logs": {
			in:      []string{"test", "test1", "test2"},
			format:  format,
			pii:     map[string]bool{"[%START_TIME%]": true},
			outLogs: []string{},
			outErr:  "",
		},
		"empty_format_string": {
			in:      []string{validLog, logWithOneField},
			format:  "",
			pii:     map[string]bool{"[%START_TIME%]": true},
			outLogs: []string{},
			outErr:  "",
		},
		"empty_pii_map": {
			in:     []string{validLog, logWithOneField},
			format: format,
			pii:    map[string]bool{},
			outLogs: []string{`'[2019-09-05T17:53:36.908Z]' 'GET /productpage HTTP/1.1' 200 - - 'Mozilla/5.0 (X11; Linux x86_64)` +
				` AppleWebKit/537.36 (KHTML, like Gecko) Chrome/76.0.3809.132 Safari/537.36'`},
			outErr: "",
		},
		"invalid_pii_map": {
			in:     []string{validLog, logWithOneField},
			format: format,
			pii:    map[string]bool{"test": true, "test2": true},
			outLogs: []string{`'[2019-09-05T17:53:36.908Z]' 'GET /productpage HTTP/1.1' 200 - - 'Mozilla/5.0 (X11; Linux x86_64)` +
				` AppleWebKit/537.36 (KHTML, like Gecko) Chrome/76.0.3809.132 Safari/537.36'`},
			outErr: "",
		},
	}

	piiFunc := func(string) string { return "pii" }

	// defining subtest with t.Run
	//nolint:gocritic,scopelint
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			filter, err := NewFilter(tc.format, piiFunc, tc.pii)
			actual := filter.Process(tc.in)

			if err != nil {
				if tc.outErr == "" {
					// expected no error but get error
					t.Fatalf("ProcessLogs(%v, %v, %v, piiFunc) = %v, %v, expected no error", tc.in, tc.format, tc.pii, actual, err)
				} else if !strings.Contains(err.Error(), tc.outErr) {
					// actual error does not contain the expected error
					t.Fatalf("ProcessLogs(%v, %v, %v, piiFunc) = %v, %v, but expected error: %v", tc.in, tc.format, tc.pii, actual, err, tc.outErr)
				}
			} else if tc.outErr != "" {
				// expect error but no acutal error
				t.Fatalf("ProcessLogs(%v, %v, %v, piiFunc) = %v, %v, but expected error: %v", tc.in, tc.format, tc.pii, actual, err, tc.outErr)
			}

			// always check if actual logs do match the expected logs
			if !reflect.DeepEqual(actual, tc.outLogs) {
				t.Fatalf("ProcessLogs(%v, %v, %v, piiFunc) = %v but expected output: %v", tc.in, tc.format, tc.pii, actual, tc.outLogs)
			}

		})
	}
}
