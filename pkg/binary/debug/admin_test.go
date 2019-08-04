package debug

import (
	"testing"

	"github.com/tetratelabs/getenvoy/pkg/binary"
	"github.com/tetratelabs/getenvoy/pkg/manifest"
)

// This test relies on a local Envoy binary, if not present it will pull one from GetEnvoy
func Test_retrieveAdminAPIData(t *testing.T) {
	tests := []struct {
		name    string
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		tc := tt
		t.Run(tc.name, func(t *testing.T) {
			key, _ := manifest.NewKey("standard/1.11.0")
			r, _ := binary.NewRuntime()
			go r.Run(key, []string{})

		})
	}
}
