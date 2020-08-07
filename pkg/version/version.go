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

package version

const (
	// defaultVersion represents a more descriptive default value for those
	// cases where a binary or unit tests get built ad-hoc without using
	// -ldflags="-s -w -X github.com/tetratelabs/getenvoy/pkg/version.version=${VERSION}"
	defaultVersion = "dev"
)

var (
	// version is populated at build time via compiler options.
	version string
)

func versionOrDefault() string {
	if version != "" {
		return version
	}
	return defaultVersion
}

// BuildInfo describes a particular build of getenvoy toolkit.
type BuildInfo struct {
	Version string
}

var (
	// Build describes a version of the enclosing binary.
	Build = BuildInfo{
		Version: versionOrDefault(),
	}
)

// IsDevBuild returns true if a version of the enclosing binary
// has not been set, which can normally happen in those case where
// the binary or unit tests get built ad-hoc.
func IsDevBuild() bool {
	return Build.Version == defaultVersion
}
