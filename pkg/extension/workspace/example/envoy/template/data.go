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

package template

import (
	"github.com/pkg/errors"

	"github.com/golang/protobuf/jsonpb"
	"github.com/golang/protobuf/proto"
	"github.com/golang/protobuf/ptypes/wrappers"

	envoycore "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
)

// newExpandData creates a template data object for the Expand operation.
func newExpandData(ctx *ExpandContext) *expandData {
	return &expandData{
		getEnvoy{
			getEnvoyExtension{
				ctx: ctx,
			},
		},
	}
}

// expandData represents a template data object that handles {{ .GetEnvoy }} pipeline.
type expandData struct {
	getEnvoy
}

// getEnvoy represents a template data object that handles {{ .GetEnvoy.DefaultValue }}
// and {{ .GetEnvoy.Extension }} pipelines.
type getEnvoy struct {
	getEnvoyExtension
}

// getEnvoyExtension represents a template data object that handles {{ .GetEnvoy.Extension.Name }},
// {{ .GetEnvoy.Extension.Code }} and {{ .GetEnvoy.Extension.Config }} pipelines.
type getEnvoyExtension struct {
	ctx *ExpandContext
}

// GetEnvoy handles {{ .GetEnvoy }} pipeline.
func (d *expandData) GetEnvoy() *getEnvoy {
	return &d.getEnvoy
}

// Extension handles {{ .GetEnvoy.Extension }} pipeline.
func (e *getEnvoy) Extension() *getEnvoyExtension {
	return &e.getEnvoyExtension
}

// Code handles {{ .GetEnvoy.Extension.Name }} pipeline.
func (e *getEnvoyExtension) Name(modules ...string) (getEnvoyValue, error) {
	if len(modules) > 0 {
		return nil, errors.Errorf("unable to resolve Wasm module %v: not supported yet", modules)
	}
	return wrap(&wrappers.StringValue{
		Value: e.ctx.DefaultExtension.GetDescriptor().Name,
	})
}

// Code handles {{ .GetEnvoy.Extension.Code }} pipeline.
func (e *getEnvoyExtension) Code(modules ...string) (getEnvoyValue, error) {
	if len(modules) > 0 {
		return nil, errors.Errorf("unable to resolve Wasm module %v: not supported yet", modules)
	}
	return wrap(&envoycore.AsyncDataSource{
		Specifier: &envoycore.AsyncDataSource_Local{
			Local: &envoycore.DataSource{
				Specifier: &envoycore.DataSource_Filename{
					Filename: e.ctx.DefaultExtension.GetWasmFile(),
				},
			},
		},
	})
}

// Config handles {{ .GetEnvoy.Extension.Config }} pipeline.
func (e *getEnvoyExtension) Config(names ...string) (getEnvoyValue, error) {
	if len(names) > 0 {
		return nil, errors.Errorf("unable to resolve a named config %v: not supported yet", names)
	}
	return wrap(&wrappers.StringValue{
		Value: e.ctx.DefaultExtensionConfig,
	})
}

// getEnvoyValue represents a value returned by various {{ .GetEnvoy.* }} pipelines.
//
// getEnvoyValue is responsible for consistent rendering of return values into JSON.
type getEnvoyValue interface {
	Message() proto.Message
	String() string
}

func wrap(message proto.Message) (getEnvoyValue, error) {
	json, err := new(jsonpb.Marshaler).MarshalToString(message)
	if err != nil {
		return nil, err
	}
	return &wrapper{message: message, json: json}, nil
}

// wrapper implements getEnvoyValue abstraction.
type wrapper struct {
	message proto.Message
	json    string
}

// Message returns the underlying Protobuf message and can be used in pipelines.
//
// E.g., {{ (.GetEnvoy.DefaultValue "admin").Message.GetAddress.GetSocketAddress.GetPortValue }}.
func (w *wrapper) Message() proto.Message {
	return w.message
}

// String returns JSON representation of the underlying Protobuf message.
func (w *wrapper) String() string {
	return w.json
}
