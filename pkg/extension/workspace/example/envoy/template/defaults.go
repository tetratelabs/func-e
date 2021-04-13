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
	"fmt"

	envoybootstrap "github.com/envoyproxy/go-control-plane/envoy/config/bootstrap/v3"
	envoycore "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

// List of properties supported by the {{ .GetEnvoy.DefaultValue "<property>" }} placeholder.
const (
	propAdmin                     = "admin"
	propAdminAccessLogPath        = "admin.access_log_path"
	propAdminAddress              = "admin.address"
	propAdminAddressSocketAddress = "admin.address.socket.address"
	propAdminAddressSocketPort    = "admin.address.socket.port"
)

var (
	defaultAdmin = func() *envoybootstrap.Admin {
		return &envoybootstrap.Admin{
			AccessLogPath: "/dev/null",
			Address: &envoycore.Address{
				Address: &envoycore.Address_SocketAddress{
					SocketAddress: &envoycore.SocketAddress{
						Address: "127.0.0.1",
						PortSpecifier: &envoycore.SocketAddress_PortValue{
							PortValue: 9901,
						},
					},
				},
			},
		}
	}
)

// DefaultValue handles {{ .GetEnvoy.DefaultValue "<property>" }} pipeline.
func (e *getEnvoy) DefaultValue(property string) (getEnvoyValue, error) {
	eval := func(property string) (proto.Message, error) {
		switch property {
		case propAdmin:
			return defaultAdmin(), nil
		case propAdminAccessLogPath:
			return &wrapperspb.StringValue{
				Value: defaultAdmin().GetAccessLogPath(),
			}, nil
		case propAdminAddress:
			return defaultAdmin().GetAddress(), nil
		case propAdminAddressSocketAddress:
			return &wrapperspb.StringValue{
				Value: defaultAdmin().GetAddress().GetSocketAddress().GetAddress(),
			}, nil
		case propAdminAddressSocketPort:
			return &wrapperspb.UInt32Value{
				Value: defaultAdmin().GetAddress().GetSocketAddress().GetPortValue(),
			}, nil
		default:
			return nil, fmt.Errorf("unknown property %q", property)
		}
	}
	value, err := eval(property)
	if err != nil {
		return nil, err
	}
	return wrap(value)
}
