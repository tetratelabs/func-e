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
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"

	"github.com/mholt/archiver/v3"
	"github.com/schollz/progressbar/v3"
	"github.com/tetratelabs/log"

	"github.com/tetratelabs/getenvoy/pkg/globals"
	"github.com/tetratelabs/getenvoy/pkg/manifest"
	"github.com/tetratelabs/getenvoy/pkg/transport"
)

const envoyLocation = "bin/envoy"

// FetchIfNeeded downloads an Envoy binary corresponding to the given reference and returns a path to it or an error.
func FetchIfNeeded(o *globals.GlobalOpts, reference string) (string, error) {
	key, e := manifest.NewKey(reference)
	if e != nil {
		return "", e
	}

	platformDirectory := filepath.Join(o.HomeDir, "builds", key.Flavor, key.Version, key.Platform)
	envoyPath := filepath.Join(platformDirectory, envoyLocation)
	stat, e := os.Stat(envoyPath)
	switch {
	case os.IsNotExist(e):
		m, err := manifest.FetchManifest(o.ManifestURL)
		if err != nil {
			return "", err
		}
		binaryLocation, err := manifest.LocateBuild(key, m)
		if err != nil {
			return "", err
		}
		if err = os.MkdirAll(platformDirectory, 0750); err != nil {
			return "", fmt.Errorf("unable to create directory %q: %w", platformDirectory, err)
		}

		log.Debugf("fetching %v from %v", key, binaryLocation)
		err = fetchEnvoy(platformDirectory, binaryLocation)
		if err != nil {
			return "", err
		}
	case e != nil:
		return "", fmt.Errorf("invalid Envoy binary at %q: %w", envoyPath, e)
	default:
		fmt.Printf("%v is already downloaded\n", key)
		if stat.Mode()&0111 == 0 {
			return "", fmt.Errorf("envoy binary not executable: %s", envoyPath)
		}
	}

	return envoyPath, nil
}

func fetchEnvoy(dst, src string) error {
	tmpDir, err := ioutil.TempDir("", "getenvoy-")
	if err != nil {
		return err
	}
	defer os.RemoveAll(tmpDir) //nolint
	tarball, err := doDownload(tmpDir, src)
	if err != nil {
		return fmt.Errorf("unable to fetch envoy from %v: %w", src, err)
	}
	err = extractEnvoy(dst, tarball)
	if err != nil {
		return fmt.Errorf("unable to extract envoy to %v: %w", dst, err)
	}
	return nil
}

func doDownload(dst, src string) (string, error) {
	// #nosec -> src can be anywhere by design
	resp, err := transport.Get(src)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close() //nolint

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("received %v status code from %q", resp.StatusCode, src)
	}

	tarball := filepath.Join(dst, "envoy.tar"+filepath.Ext(src))
	// #nosec -> dst can be anywhere by design
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
						return fmt.Errorf("error extracting %v: %w", f.Name(), err)
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
