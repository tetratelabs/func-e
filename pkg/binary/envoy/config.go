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

package envoy

import (
	"fmt"
	"time"

	"github.com/gogo/protobuf/types"
)

type Mode string

const (
	Sidecar Mode = "sidecar"
	Router  Mode = "router"
)

var SupportedModes = []string{string(Router)}

func ParseMode(s string) (Mode, error) {
	switch {
	case Mode(s) == Sidecar:
		return Sidecar, nil
	case Mode(s) == Router:
		return Router, nil
	case s == "":
		return "", nil
	default:
		return "", fmt.Errorf("unable to parse mode %v, must be one of %v", s, SupportedModes)
	}
}

func NewConfig(options ...func(*Config)) *Config {
	cfg := &Config{
		AdminPort:      15000,
		StatNameLength: 189,
		DrainDuration:  types.DurationProto(30 * time.Second),
		ConnectTimeout: types.DurationProto(5 * time.Second),
	}
	for _, o := range options {
		o(cfg)
	}
	return cfg
}

type Config struct {
	XDSAddress     string
	Mode           Mode
	IPAddresses    []string
	ALSAddresss    string
	DrainDuration  *types.Duration
	ConnectTimeout *types.Duration
	AdminPort      int32
	StatNameLength int32
}
