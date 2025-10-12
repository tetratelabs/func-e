// Copyright 2025 Tetrate
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestParseLogLevel(t *testing.T) {
	tests := []struct {
		input    string
		expected logLevel
	}{
		{"trace", logLevelTrace},
		{"TRACE", logLevelTrace},
		{"debug", logLevelDebug},
		{"Debug", logLevelDebug},
		{"info", logLevelInfo},
		{"INFO", logLevelInfo},
		{"warning", logLevelWarning},
		{"warn", logLevelWarning},
		{"WARN", logLevelWarning},
		{"error", logLevelError},
		{"ERROR", logLevelError},
		{"critical", logLevelCritical},
		{"CRITICAL", logLevelCritical},
		{"off", logLevelOff},
		{"OFF", logLevelOff},
		{"unknown", logLevelInfo}, // defaults to info
		{"", logLevelInfo},        // defaults to info
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			actual := parseLogLevel(tt.input)
			require.Equal(t, tt.expected, actual)
		})
	}
}

func TestLogLevelOrdering(t *testing.T) {
	// Verify that log levels are ordered correctly for comparison
	require.Less(t, logLevelTrace, logLevelDebug)
	require.Less(t, logLevelDebug, logLevelInfo)
	require.Less(t, logLevelInfo, logLevelWarning)
	require.Less(t, logLevelWarning, logLevelError)
	require.Less(t, logLevelError, logLevelCritical)
	require.Less(t, logLevelCritical, logLevelOff)
}
