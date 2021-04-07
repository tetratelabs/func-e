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

package util

import (
	"reflect"
	"strings"
	"testing"

	envoylistener "github.com/envoyproxy/go-control-plane/envoy/config/listener/v3"
	"github.com/golang/protobuf/proto"
	"github.com/stretchr/testify/require"

	"github.com/tetratelabs/getenvoy/pkg/extension/workspace/model"
)

func TestLoadAllowsUnknownAnyTypes(t *testing.T) {
	config := &model.File{Source: "/path/to/envoy.yaml", Content: []byte(`
        unknown_field: unknown_value
        filter_chains:
        - filters:
          - name: example
            typed_config:
              "@type": "type.googleapis.com/unknown.Type"
              a: b
              c: 1000
`)}

	var listener envoylistener.Listener
	err := Load(config, &listener)
	require.NoError(t, err)
}

func TestLoadValidatesYAML(t *testing.T) {
	config := &model.File{Source: "/path/to/envoy.yaml", Content: []byte(`
        code: {{ .GetEnvoy.Extension.Code }}
`)}

	var listener envoylistener.Listener
	err := Load(config, &listener)

	expectedErr := `failed to convert into JSON Envoy config coming from "/path/to/envoy.yaml": yaml: invalid map key: map[interface {}]interface {}{".GetEnvoy.Extension.Code":interface {}(nil)}`
	require.EqualError(t, err, expectedErr)
}

func TestNewFakeAnyResolver(t *testing.T) {
	tests := []struct {
		name    string
		typeURL string
	}{
		{"unknown type", "type.googleapis.com/unknown.Type"},
		{"known type", "type.googleapis.com/envoy.config.bootstrap.v3.Bootstrap"},
	}

	for _, test := range tests {
		test := test // pin! see https://github.com/kyoh86/scopelint for why

		t.Run(test.name, func(t *testing.T) {
			resolver := newFakeAnyResolver()

			actual, err := resolver.Resolve(test.typeURL)
			require.NoError(t, err)

			props := proto.GetProperties(reflect.ValueOf(actual).Elem().Type())
			fields := MessageFields(props.Prop).Filter(func(field *proto.Properties) bool {
				return !strings.HasPrefix(field.Name, "XXX_")
			})
			require.Empty(t, fields)
		})
	}
}

type MessageFields []*proto.Properties

func (fields MessageFields) Filter(test func(*proto.Properties) bool) MessageFields {
	result := make(MessageFields, 0)
	for _, field := range fields {
		if test(field) {
			result = append(result, field)
		}
	}
	return result
}
