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

package cmd_test

import (
	"errors"

	"github.com/spf13/cobra"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	. "github.com/tetratelabs/getenvoy/pkg/util/cmd"
)

var _ = Describe("ComposableRunE", func() {
	Describe("ThenE()", func() {
		It("should not call the next func if a previous one fails", func() {
			expectedErr := errors.New("expected")
			unexpectedErr := errors.New("unexpected")

			compositeFn := ComposableRunE(func(cmd *cobra.Command, args []string) error {
				return expectedErr
			}).ThenE(func(cmd *cobra.Command, args []string) error {
				return unexpectedErr
			})

			actualErr := compositeFn(new(cobra.Command), nil)

			Expect(actualErr).To(Equal(expectedErr))
		})

		It("should call the next func if a previous one doesn't fail", func() {
			expectedErr := errors.New("expected")

			compositeFn := ComposableRunE(func(cmd *cobra.Command, args []string) error {
				return nil
			}).ThenE(func(cmd *cobra.Command, args []string) error {
				return expectedErr
			})

			actualErr := compositeFn(new(cobra.Command), nil)

			Expect(actualErr).To(Equal(expectedErr))
		})
	})

	Describe("Then()", func() {
		It("should not call the next func if a previous one fails", func() {
			expectedErr := errors.New("expected")
			nextCalled := false

			compositeFn := ComposableRunE(func(cmd *cobra.Command, args []string) error {
				return expectedErr
			}).Then(func(cmd *cobra.Command, args []string) {
				nextCalled = true
			})

			actualErr := compositeFn(new(cobra.Command), nil)

			Expect(actualErr).To(Equal(expectedErr))
			Expect(nextCalled).To(BeFalse())
		})

		It("should call the next func if a previous one doesn't fail", func() {
			nextCalled := false

			compositeFn := ComposableRunE(func(cmd *cobra.Command, args []string) error {
				return nil
			}).Then(func(cmd *cobra.Command, args []string) {
				nextCalled = true
			})

			actualErr := compositeFn(new(cobra.Command), nil)

			Expect(actualErr).To(BeNil())
			Expect(nextCalled).To(BeTrue())
		})
	})
})

var _ = Describe("CallParentPersistentPreRunE()", func() {
	It("should call PersistentPreRunE on the parent", func() {
		expectedErr := errors.New("expected")
		root := &cobra.Command{
			PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
				return expectedErr
			},
		}
		nested := new(cobra.Command)
		root.AddCommand(nested)

		actualErr := CallParentPersistentPreRunE()(nested, nil)

		Expect(actualErr).To(Equal(expectedErr))
	})

	It("should call PersistentPreRun on the parent", func() {
		parentCalled := false
		root := &cobra.Command{
			PersistentPreRun: func(cmd *cobra.Command, args []string) {
				parentCalled = true
			},
		}
		nested := new(cobra.Command)
		root.AddCommand(nested)

		actualErr := CallParentPersistentPreRunE()(nested, nil)

		Expect(actualErr).To(BeNil())
		Expect(parentCalled).To(BeTrue())
	})
})
