package api

import (
	"bytes"
	"context"
	"os"
	"path/filepath"

	"github.com/stretchr/testify/require"

	"testing"
)

var (
	// minRunArgs is the minimal config needed to run Envoy 1.18+, non-windows <1.18 need access_log_path: '/dev/stdout'
	minRunArgs = []string{"--config-yaml", "admin: {address: {socket_address: {address: '127.0.0.1', port_value: 0}}}"}
)

func TestRunWithHomeDir(t *testing.T) {
	ctx := context.Background()
	tmpDir := t.TempDir()
	require.NoError(t, Run(ctx, minRunArgs, HomeDir(tmpDir)))
	_, err := os.Stat(filepath.Join(tmpDir, "versions"))
	require.NoError(t, err)
}

func TestRunWithOut(t *testing.T) {
	ctx := context.Background()
	tmpDir := t.TempDir()
	b := bytes.NewBufferString("")
	require.Equal(t, 0, b.Len())
	require.NoError(t, Run(ctx, minRunArgs, Out(b), HomeDir(tmpDir)))
	require.NotEqual(t, 0, b.Len())
}
