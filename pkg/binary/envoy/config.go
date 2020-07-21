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
	"os"
	"path/filepath"
	"time"

	"github.com/gogo/protobuf/types"
)

// Mode is the mode Envoy should run in
type Mode string

const (
	// Sidecar instructs Envoy to run as a sidecar
	Sidecar Mode = "sidecar"
	// LoadBalancer instructs Envoy to run as a loadbalancer (e.g. gateway)
	LoadBalancer Mode = "loadbalancer"
)

// SupportedModes indicate the modes that are current supported by GetEnvoy
var SupportedModes = []string{string(LoadBalancer)}

// ParseMode converts the passed string into a valid mode or empty string
func ParseMode(s string) Mode {
	switch Mode(s) {
	case Sidecar:
		return Sidecar
	case LoadBalancer:
		return LoadBalancer
	default:
		return ""
	}
}

// NewConfig creates and mutates a config object based on passed params
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

// Config store Envoy config information for use by bootstrap and arg mutators
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

func (r *Runtime) SaveConfig(name string, config string) (string, error) {
	configDir := filepath.Join(r.RootDir, "configs")
	if err := os.MkdirAll(configDir, 0750); err != nil {
		return "", fmt.Errorf("unable to create directory %q: %v", configDir, err)
	}

	f, err := os.OpenFile(filepath.Join(configDir, name+".yaml"), os.O_RDWR|os.O_CREATE, 0660)
	if err != nil {
		return "", fmt.Errorf("Cannot create config directory %s: %s", configDir, err)
	}

	defer f.Close()

	if _, err = f.WriteString(config); err != nil {
		return "", fmt.Errorf("Cannot save config file %s: %s", filepath.Join(configDir, name+".yaml"), err)
	}

	return filepath.Join(configDir, name+".yaml"), nil
}
