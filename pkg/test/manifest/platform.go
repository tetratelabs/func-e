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

package manifest

import (
	"fmt"
	"strings"

	"github.com/tetratelabs/getenvoy/api"
)

// Platform represents a platform.
type Platform string

const (
	// PlatformLinuxGlibc represents a `linux-glibc` platform.
	PlatformLinuxGlibc = "linux-glibc"
	// PlatformDarwin represents a `darwin` platform.
	PlatformDarwin = "darwin"
)

func (p Platform) String() string {
	return string(p)
}

// BuildPlatform returns an equivalent Build_Platform.
func (p Platform) BuildPlatform() api.Build_Platform {
	switch p {
	case PlatformLinuxGlibc:
		return api.Build_LINUX_GLIBC
	case PlatformDarwin:
		return api.Build_DARWIN
	default:
		panic(fmt.Errorf("unknown platform %q", p))
	}
}

// Code returns an equivalent platform code name.
func (p Platform) Code() string {
	code := strings.ToUpper(string(p))
	code = strings.ReplaceAll(code, "-", "_")
	return code
}

// Platforms represents a list of platforms.
type Platforms []Platform

// Index return map of platforms indexed by string value.
func (ps Platforms) Index() map[string]Platform {
	index := make(map[string]Platform, len(ps))
	for _, p := range ps {
		index[string(p)] = p
	}
	return index
}

var (
	// SupportedPlatforms represents a list of all supported platforms.
	SupportedPlatforms = Platforms{PlatformLinuxGlibc, PlatformDarwin}
)

var (
	supportedPlatformsIndex = SupportedPlatforms.Index()
)

// ParsePlatform parses a string representation of a platform.
func ParsePlatform(text string) (Platform, error) {
	p, ok := supportedPlatformsIndex[text]
	if !ok {
		return "", fmt.Errorf("unknown platform %q", text)
	}
	return p, nil
}
