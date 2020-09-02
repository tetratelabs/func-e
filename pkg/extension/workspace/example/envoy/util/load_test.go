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

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"

	"github.com/golang/protobuf/proto"

	envoylistener "github.com/envoyproxy/go-control-plane/envoy/config/listener/v3"

	"github.com/tetratelabs/getenvoy/pkg/extension/workspace/model"
)

var _ = Describe("Load()", func() {
	It("should not fail on unknown Any types", func() {
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
		Expect(err).ToNot(HaveOccurred())
	})

	//nolint:lll
	It("should fail if input is not a valid YAML", func() {
		config := &model.File{Source: "/path/to/envoy.yaml", Content: []byte(`
        code: {{ .GetEnvoy.Extension.Code }}
`)}
		expectedErr := `failed to convert into JSON Envoy config coming from "/path/to/envoy.yaml": yaml: invalid map key: map[interface {}]interface {}{".GetEnvoy.Extension.Code":interface {}(nil)}`

		var listener envoylistener.Listener
		err := Load(config, &listener)
		Expect(err).To(MatchError(expectedErr))
	})
})

var _ = Describe("newFakeAnyResolver()", func() {
	DescribeTable("should be able to resolve any type URL",
		func(given string) {
			resolver := newFakeAnyResolver()

			actual, err := resolver.Resolve(given)
			Expect(err).ToNot(HaveOccurred())

			props := proto.GetProperties(reflect.ValueOf(actual).Elem().Type())
			fields := MessageFields(props.Prop).Filter(func(field *proto.Properties) bool {
				return !strings.HasPrefix(field.Name, "XXX_")
			})
			Expect(fields).To(HaveLen(0))
		},
		Entry("unknown type", "type.googleapis.com/unknown.Type"),
		Entry("known type", "type.googleapis.com/envoy.config.bootstrap.v3.Bootstrap"),
	)
})

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
