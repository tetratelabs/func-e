package api

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/tetratelabs/func-e/internal/test"
	"github.com/tetratelabs/func-e/internal/version"
)

var (
	// minRunArgs is the minimal config needed to run Envoy 1.18+, non-windows <1.18 need access_log_path: '/dev/stdout'
	minRunArgs = []string{"--config-yaml", "admin: {address: {socket_address: {address: '127.0.0.1', port_value: 0}}}"}
)

func TestRun(t *testing.T) {
	ctx := context.Background()
	tmpDir := t.TempDir()
	envoyVersion := version.LastKnownEnvoy
	versionsServer := test.RequireEnvoyVersionsTestServer(t, envoyVersion)
	defer versionsServer.Close()
	envoyVersionsURL := versionsServer.URL + "/envoy-versions.json"
	b := bytes.NewBufferString("")

	require.Equal(t, 0, b.Len())
	require.NoError(t, Run(ctx, minRunArgs, Out(b), HomeDir(tmpDir), EnvoyVersionsURL(envoyVersionsURL)))
	require.NotEqual(t, 0, b.Len())
	_, err := os.Stat(filepath.Join(tmpDir, "versions"))
	require.NoError(t, err)
}
