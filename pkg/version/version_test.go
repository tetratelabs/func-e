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

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
)

var _ = Describe("versionOrDefault()", func() {

	var backupVersion string

	BeforeEach(func() {
		backupVersion = version
	})

	AfterEach(func() {
		version = backupVersion
	})

	type testCase struct {
		version  string
		expected string
	}

	DescribeTable("should fallback to `dev` version if `pkg/version.version` was not set via compiler options",
		func(given testCase) {
			version = given.version

			Expect(versionOrDefault()).To(Equal(given.expected))
		},
		Entry("ad-hoc build", testCase{
			version:  "",
			expected: "dev",
		}),
		Entry("dev build", testCase{
			version:  "dev",
			expected: "dev",
		}),
		Entry("release build", testCase{
			version:  "0.0.1",
			expected: "0.0.1",
		}),
	)
})

var _ = Describe("IsDevBuild()", func() {

	var backupBuild BuildInfo

	BeforeEach(func() {
		backupBuild = Build
	})

	AfterEach(func() {
		Build = backupBuild
	})

	type testCase struct {
		build    BuildInfo
		expected bool
	}

	DescribeTable("should consider builds with `pkg/version.version` unset or set to `dev` as 'development builds'",
		func(given testCase) {
			Build = given.build

			Expect(IsDevBuild()).To(Equal(given.expected))
		},
		Entry("dev build", testCase{
			build:    BuildInfo{Version: "dev"},
			expected: true,
		}),
		Entry("release build", testCase{
			build:    BuildInfo{Version: "0.0.1"},
			expected: false,
		}),
	)
})
