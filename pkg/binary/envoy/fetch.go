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

package envoy

import (
	"archive/tar"
	"errors"
	"net/http"
	"path/filepath"
	"strings"

	"os"

	"fmt"

	"io/ioutil"

	"io"

	"github.com/mholt/archiver"
	"github.com/schollz/progressbar/v2"
	"github.com/tetratelabs/getenvoy/pkg/manifest"
	"github.com/tetratelabs/log"
)

const envoyLocation = "bin/envoy"

// FetchAndRun downloads an Envoy binary, if necessary, and runs it.
func (r *Runtime) FetchAndRun(reference string, args []string) error {
	key, err := manifest.NewKey(reference)
	if err != nil {
		if _, err := os.Stat(reference); err != nil {
			return fmt.Errorf("%q is neither a valid Envoy release provided by getenvoy.io nor a path to a custom Envoy binary", reference)
		}
		return r.RunPath(reference, args)
	}
	if !r.AlreadyDownloaded(key) {
		location, err := manifest.Locate(key)
		if err != nil {
			return err
		}
		if err := r.Fetch(key, location); err != nil {
			return err
		}
	}
	return r.Run(key, args)
}

// Fetch downloads an Envoy binary from the passed location
func (r *Runtime) Fetch(key *manifest.Key, binaryLocation string) error {
	if !r.AlreadyDownloaded(key) {
		log.Debugf("fetching %v from %v", key, binaryLocation)
		dst := r.platformDirectory(key)
		if err := os.MkdirAll(dst, 0750); err != nil {
			return fmt.Errorf("unable to create directory %q: %v", dst, err)
		}
		fmt.Printf("fetching %v\n", key)
		return fetchEnvoy(dst, binaryLocation)
	}
	fmt.Printf("%v is already downloaded\n", key)
	return nil
}

// AlreadyDownloaded returns true if there is a cached Envoy binary matching the passed Key
func (r *Runtime) AlreadyDownloaded(key *manifest.Key) bool {
	_, err := os.Stat(filepath.Join(r.platformDirectory(key), envoyLocation))

	// !IsNotExist is not the same as IsExist
	// os.Stat doesn't return IsExist typed errors
	return !os.IsNotExist(err)
}

// BinaryStore returns the location at which the runtime instance persists binaries
// Getters typically aren't idiomatic Go, however, this one is deliberately part of the fetcher interface
func (r *Runtime) BinaryStore() string {
	return filepath.Join(r.store, "builds")
}

func (r *Runtime) platformDirectory(key *manifest.Key) string {
	platform := strings.ToLower(key.Platform)
	platform = strings.ReplaceAll(platform, "-", "_")
	return filepath.Join(r.BinaryStore(), key.Flavor, key.Version, platform)
}

func fetchEnvoy(dst, src string) error {
	tmpDir, err := ioutil.TempDir("", "getenvoy-")
	if err != nil {
		return err
	}
	defer os.RemoveAll(tmpDir) //nolint
	tarball, err := doDownload(tmpDir, src)
	if err != nil {
		return fmt.Errorf("unable to fetch envoy from %v: %v", src, err)
	}
	if err := extractEnvoy(dst, tarball); err != nil {
		return fmt.Errorf("unable to extract envoy to %v: %v", dst, err)
	}
	return nil
}

func doDownload(dst, src string) (string, error) {
	// #nosec -> src destination can be anywhere by design
	resp, err := http.Get(src)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close() //nolint

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("received %v status code from %q", resp.StatusCode, src)
	}

	tarball := filepath.Join(dst, "envoy.tar"+filepath.Ext(src))
	f, err := os.OpenFile(tarball, os.O_CREATE|os.O_WRONLY, 0600)
	if err != nil {
		return "", err
	}
	defer f.Close() //nolint

	bar := progressbar.NewOptions64(resp.ContentLength, progressbar.OptionSetDescription("[Fetching Envoy]"))
	out := io.MultiWriter(f, bar)
	_, err = io.Copy(out, resp.Body)
	fmt.Println("") // append a newline to progressbar output
	return tarball, err
}

func extractEnvoy(dst, tarball string) error {
	// Walk the tarball until we find the bin and lib directories
	if err := archiver.Walk(tarball, func(f archiver.File) error {
		if (f.Name() == "bin" && f.IsDir()) || (f.Name() == "lib" && f.IsDir()) {
			if f.Header != nil {
				if header, ok := f.Header.(*tar.Header); ok {
					if err := archiver.Extract(tarball, header.Name, dst); err != nil {
						log.Errorf("error extracting %v: %v", f.Name(), err)
					}
				}
			}
		}
		return nil
	}); err != nil {
		return err
	}
	envoyFilepath := filepath.Join(dst, envoyLocation)
	log.Debugf("checking for binary at %v", envoyFilepath)
	if _, err := os.Stat(envoyFilepath); os.IsNotExist(err) {
		return errors.New("no Envoy binary in downloaded tarball")
	}
	return nil
}
