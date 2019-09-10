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
	"testing"

	"bitbucket.org/creachadair/shell"
	"github.com/google/go-cmp/cmp"
)

func TestProcess(t *testing.T) {
	var FORMAT_STRING string = "[%START_TIME%] \"%REQ(:METHOD)% %REQ(X-ENVOY-ORIGINAL-PATH?:PATH)% %PROTOCOL%\" %RESPONSE_CODE% %RESPONSE_FLAGS% \"%DYNAMIC_METADATA(istio.mixer:status)%\" \"%UPSTREAM_TRANSPORT_FAILURE_REASON%\" %BYTES_RECEIVED% %BYTES_SENT% %DURATION% %RESP(X-ENVOY-UPSTREAM-SERVICE-TIME)% \"%REQ(X-FORWARDED-FOR)%\" \"%REQ(USER-AGENT)%\" \"%REQ(X-REQUEST-ID)%\" \"%REQ(:AUTHORITY)%\" \"%UPSTREAM_HOST%\" %UPSTREAM_CLUSTER% %UPSTREAM_LOCAL_ADDRESS% %DOWNSTREAM_LOCAL_ADDRESS% %DOWNSTREAM_REMOTE_ADDRESS% %REQUESTED_SERVER_NAME%\n"

	FORMATS, ok := shell.Split(FORMAT_STRING)
	if !ok {
		t.Fatalf("error in splitting the FORMAT_STRING: %s", FORMAT_STRING)
	}

	var VALID_LOG string = `[2019-09-05T17:53:36.908Z] "GET /productpage HTTP/1.1" 200 - "-" "-" 0 5183 47 46 "10.28.2.1" "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/76.0.3809.132 Safari/537.36" "4273c09c-2534-962a-b7da-04a04cf39e88" "35.199.144.198" "127.0.0.1:9080" inbound|9080|http|productpage.default.svc.cluster.local - 10.28.2.27:9080 10.28.2.1:0 outbound_.9080_._.productpage.default.svc.cluster.local`
	var LOG_WITH_ONE_FIELD string = `[2019-09-05T17:53:36.908Z]`

	type Out struct {
		logs []string
		err  error
	}
	// test cases wrapped up in a map where key is the test case name and value is the struct
	// defining the inputs and expecte outcome
	tests := map[string]struct {
		logs        []string
		format      string
		containsPII map[string]bool
		want        Out
	}{
		// positive test cases
		// only return an array containing one valid log with [%START_TIME%] field hashed
		"simple": {logs: []string{VALID_LOG, LOG_WITH_ONE_FIELD},
			format: FORMAT_STRING, containsPII: map[string]bool{"[%START_TIME%]": true},
			want: Out{logs: []string{VALID_LOG}, err: nil},
		},

		// negative test cases
		"empty_logs": {logs: []string{},
			format: FORMAT_STRING, containsPII: map[string]bool{"[%START_TIME%]": true},
			want: Out{logs: []string{}, err: nil},
		},
		"empty_format_string": {logs: []string{VALID_LOG, LOG_WITH_ONE_FIELD},
			format: "", containsPII: map[string]bool{"[%START_TIME%]": true},
			want: Out{logs: []string{}, err: nil},
		},
		"empty_pii_map": {logs: []string{VALID_LOG, LOG_WITH_ONE_FIELD},
			format: FORMAT_STRING, containsPII: map[string]bool{},
			want: Out{logs: []string{VALID_LOG}, err: nil},
		},
	}

	// defining subtest with t.Run
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			actualLogs, actualErr := Process_modified(tc.logs, tc.format, tc.containsPII)

			// hash the desired fields for each log in wanted logs
			for i, log := range tc.want.logs {
				fields, ok := shell.Split(log)
				if !ok {
					t.Fatalf("error in splitting the log: %s", log)
				}

				for j, fieldName := range FORMATS {
					if tc.containsPII[fieldName] {
						h, err := Hash(fields[j])
						if err != nil {
							t.Fatalf("error in hashing expected field")
						}
						fields[j] = h
					}
				}
				tc.want.logs[i] = shell.Join(fields)
				//tc.want.logs[i] = "'" + log + "'" // adding single quote to the test case
			}
			errorDiff := cmp.Diff(tc.want.err, actualErr)
			logsDiff := cmp.Diff(tc.want.logs, actualLogs)

			if errorDiff != "" || logsDiff != "" {
				t.Fatal(logsDiff)
				t.Fatalf(errorDiff)
			}
		})
	}
}
