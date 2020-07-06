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

package version_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"

	. "github.com/tetratelabs/getenvoy/pkg/version"
)

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

	DescribeTable("",
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
