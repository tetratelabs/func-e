package envoytest

import (
	"fmt"

	"github.com/tetratelabs/getenvoy/pkg/binary/envoy"
	"github.com/tetratelabs/getenvoy/pkg/manifest"
)

const Reference = "standard:1.11.0"

func Fetch() error {
	key, _ := manifest.NewKey(Reference)
	r, _ := envoy.NewRuntime()
	if !r.AlreadyDownloaded(key) {
		location, err := manifest.Locate(key, manifest.DefaultURL)
		if err != nil {
			return fmt.Errorf("unable to retrieve manifest from %v: %v", manifest.DefaultURL, err)
		}
		if err := r.Fetch(key, location); err != nil {
			return fmt.Errorf("unable to retrieve binary from %v: %v", location, err)
		}
	}
	return nil
}
