package debug

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"

	"github.com/hashicorp/go-multierror"
	"github.com/tetratelabs/getenvoy/pkg/binary"
)

var adminAPIPaths = map[string]string{
	"certs":             "certs.json",
	"clusters":          "clusters.txt",
	"config_dump":       "config_dump.json",
	"contention":        "contention.txt",
	"listeners":         "listeners.txt",
	"memory":            "memory.json",
	"server_info":       "server_info.json",
	"stats?format=json": "stats.json",
	"runtime":           "runtime.json",
}

// EnableEnvoyAdminDataCollection is a preset option that registers collection of Envoy Admin API information
var EnableEnvoyAdminDataCollection = func(r *binary.Runtime) {
	r.RegisterPreTermination(retrieveAdminAPIData)
}

func retrieveAdminAPIData(r *binary.Runtime) error {
	var multiErr *multierror.Error
	for path, file := range adminAPIPaths {
		resp, err := http.Get(fmt.Sprintf("http://localhost:15001/%v", path))
		if err != nil {
			multiErr = multierror.Append(multiErr, err)
		}
		f, err := os.OpenFile(filepath.Join(r.DebugDir, file), os.O_CREATE|os.O_WRONLY, 0600)
		if err != nil {
			multiErr = multierror.Append(multiErr, err)
		}
		defer func() { _ = f.Close() }()
		defer func() { _ = resp.Body.Close() }()
		if _, err := io.Copy(f, resp.Body); err != nil {
			multiErr = multierror.Append(multiErr, err)
		}
	}
	return multiErr.ErrorOrNil()
}
