// Copyright 2021 Tetrate
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

package main

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/tetratelabs/func-e/internal/moreos"
)

func TestRunErrors(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(200)
	}))
	t.Cleanup(func() { server.Close() })

	tests := []struct {
		name           string
		args           []string
		expectedStatus int
		expectedStdout string
		expectedStderr string
	}{
		{
			name: "built-in --version output",
			args: []string{"func-e", "--version"},
			expectedStdout: `func-e version dev
`,
		},
		{
			name:           "incorrect global flag name",
			args:           []string{"func-e", "--d"},
			expectedStatus: 1,
			expectedStderr: `flag provided but not defined: -d
show usage with: func-e help
`,
		},
		{
			name:           "incorrect global flag value",
			args:           []string{"func-e", "--envoy-versions-url", ".", "list"},
			expectedStatus: 1,
			expectedStderr: `"." is not a valid Envoy versions URL
show usage with: func-e help
`,
		},
		{
			name:           "unknown command",
			args:           []string{"func-e", "fly"},
			expectedStatus: 1,
			expectedStderr: `unknown command "fly"
show usage with: func-e help
`,
		},
		{
			name:           "execution error",
			args:           []string{"func-e", "--envoy-versions-url", server.URL, "versions", "-a"},
			expectedStatus: 1,
			expectedStderr: `error: error unmarshalling Envoy versions: unexpected end of JSON input
`,
		},
	}

	for _, test := range tests {
		test := test // pin! see https://github.com/kyoh86/scopelint for why

		t.Run(test.name, func(t *testing.T) {
			stdout := new(bytes.Buffer)
			stderr := new(bytes.Buffer)

			status := run(stdout, stderr, test.args)
			require.Equal(t, test.expectedStatus, status)
			require.Equal(t, moreos.Sprintf(test.expectedStdout), stdout.String())
			require.Equal(t, moreos.Sprintf(test.expectedStderr), stderr.String())
		})
	}
}
