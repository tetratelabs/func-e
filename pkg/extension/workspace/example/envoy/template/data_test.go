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

package template

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func defaultExtensionConfigJSON(t *testing.T, defaultExtensionConfig string) string {
	e := getEnvoyExtension{ctx: &ExpandContext{DefaultExtensionConfig: defaultExtensionConfig}}

	getEnvoyValue, err := e.Config()
	require.NoError(t, err)
	require.NotNil(t, getEnvoyValue)

	return getEnvoyValue.String()
}

func TestConfigMarshalsDefaultExtensionConfig(t *testing.T) {
	json := defaultExtensionConfigJSON(t, `{"key":"value"}`)

	// Google made json formatting (json.prepareNext) intentionally unstable, technically by adding a space randomly.
	// https://github.com/golang/protobuf/issues/920 requested an option for stability, but it was closed and locked.
	// https://github.com/golang/protobuf/issues/1121 remains open, but unlikely to change.
	// Hence, we have to check two possible formats.
	withoutSpace := `{"@type":"type.googleapis.com/google.protobuf.StringValue","value":"{\"key\":\"value\"}"}`
	withSpace := `{"@type":"type.googleapis.com/google.protobuf.StringValue", "value":"{\"key\":\"value\"}"}`
	if withoutSpace != json {
		require.Equal(t, withSpace, json)
	}
}
