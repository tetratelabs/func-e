// Copyright func-e contributors
// SPDX-License-Identifier: Apache-2.0

package lint

// Pins go-ruleguard/dsl so that go mod tidy keeps it for .golangci.ruleguard.go.
import _ "github.com/quasilyte/go-ruleguard/dsl"
