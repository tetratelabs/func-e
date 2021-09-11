package test

import (
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/tetratelabs/func-e/internal/moreos"
)

var (
	_, currentFilePath, _, _ = runtime.Caller(0)

	// funcEGoModuleDir points to the directory where the func-e's go.mod resides.
	funcEGoModuleDir = filepath.Join(filepath.Dir(currentFilePath), "..", "..")
)

type fakeBinarySrc struct {
	content []byte
	path    string
}

// requireBuildFakeBinary builds a fake binary from source, either from its content or path.
func requireBuildFakeBinary(t *testing.T, name string, binarySrc fakeBinarySrc) []byte {
	goBin := requireGoBin(t)
	tempDir := t.TempDir()
	buildDir := funcEGoModuleDir
	bin := filepath.Join(tempDir, name+moreos.Exe)
	src := binarySrc.path
	if src == "" {
		buildDir = tempDir
		src = name + ".go"
		require.NoError(t, os.WriteFile(filepath.Join(tempDir, src), binarySrc.content, 0600))
	}
	cmd := exec.Command(goBin, "build", "-o", bin, src) //nolint:gosec
	cmd.Dir = buildDir
	out, err := cmd.CombinedOutput()
	require.NoError(t, err, "couldn't compile %s: %s", src, string(out))
	bytes, err := os.ReadFile(bin)
	require.NoError(t, err)
	return bytes
}

func requireGoBin(t *testing.T) string {
	binName := "go" + moreos.Exe
	goBin := filepath.Join(runtime.GOROOT(), "bin", binName)
	if _, err := os.Stat(goBin); err == nil {
		return goBin
	}
	// Now, search the path
	goBin, err := exec.LookPath(binName)
	require.NoError(t, err, "couldn't find %s in the PATH", goBin)
	return goBin
}
