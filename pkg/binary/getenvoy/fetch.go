// Copyright 2019 Tetrate
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

package getenvoy

import (
	"errors"
	"net/http"
	"path/filepath"

	"os"

	"fmt"

	"io/ioutil"

	"io"

	"github.com/mholt/archiver"
	"github.com/schollz/progressbar/v2"
	"github.com/tetratelabs/getenvoy/pkg/manifest"
)

// Fetch downloads an Envoy binary from the passed location
func (r *Runtime) Fetch(key *manifest.Key, binaryLocation string) error {
	dst := filepath.Join(r.local, key.Flavor, key.Version, key.Platform)
	if err := os.MkdirAll(dst, 0750); err != nil {
		return fmt.Errorf("unable to create directory %q: %v", dst, err)
	}
	return fetchEnvoy(dst, binaryLocation)
}

func fetchEnvoy(dst, src string) error {
	tmpDir, err := ioutil.TempDir("", "getenvoy-")
	if err != nil {
		return err
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	tarball, err := doDownload(tmpDir, src)
	if err != nil {
		return err
	}
	return extractEnvoy(dst, tarball)
}

func doDownload(dst, src string) (string, error) {
	// #nosec -> src destination can be anywhere by design
	resp, err := http.Get(src)
	if err != nil {
		return "", err
	}
	defer func() { _ = resp.Body.Close() }()

	tarball := filepath.Join(dst, "envoy.tar.gz")
	f, err := os.OpenFile(tarball, os.O_CREATE|os.O_WRONLY, 0600)
	if err != nil {
		return "", err
	}
	defer func() { _ = f.Close() }()

	bar := progressbar.NewOptions64(resp.ContentLength, progressbar.OptionSetDescription("[Fetching Envoy]"))
	out := io.MultiWriter(f, bar)
	_, err = io.Copy(out, resp.Body)
	return tarball, err
}

func extractEnvoy(dst, tarball string) error {
	// #nosec -> envoy binary needs to be executable
	envoy, err := os.OpenFile(filepath.Join(dst, "envoy"), os.O_CREATE|os.O_WRONLY, 0755)
	if err != nil {
		return err
	}
	defer func() { _ = envoy.Close() }()
	// Walk the tarball until we find file named envoy, then copy to our destination
	found := false
	if err := archiver.Walk(tarball, func(f archiver.File) error {
		if f.Name() == "envoy" && !f.IsDir() {
			found = true
			if _, err := io.Copy(envoy, f); err != nil {
				return err
			}
		}
		return nil
	}); err != nil {
		return err
	}
	if !found {
		return errors.New("unable to find Envoy binary in downloaded tarball")
	}
	return nil
}
