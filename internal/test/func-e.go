package test

import (
	"os"
	"path/filepath"
	"sync"
	"testing"

	"github.com/stretchr/testify/require"
)

var (
	// fakeFuncEBin is the compiled code of fakeEnvoySrc which will be runtime.GOOS dependent.
	fakeFuncEBin     []byte
	builtFakeFuncE   sync.Once
	fakeFuncESrcPath = filepath.Join(funcEGoModuleDir, "internal", "test", "testdata", "fake_func-e", "fake_func-e.go")
)

// RequireFakeEnvoy writes fakeEnvoyBin to the given path
func RequireFakeFuncE(t *testing.T, path string) {
	builtFakeFuncE.Do(func() {
		fakeFuncEBin = requireBuildFakeFuncE(t)
	})
	require.NoError(t, os.WriteFile(path, fakeFuncEBin, 0700)) //nolint:gosec
}

// requireBuildFakeEnvoy builds a fake envoy binary and returns its contents.
func requireBuildFakeFuncE(t *testing.T) []byte {
	return requireBuildFakeBinary(t, "func-e", fakeBinarySrc{path: fakeFuncESrcPath})
}
