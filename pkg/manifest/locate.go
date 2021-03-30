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

package manifest

import (
	"fmt"
	"net/url"
	"os"

	"github.com/pkg/errors"
	"github.com/tetratelabs/log"

	"github.com/tetratelabs/getenvoy-package/api"
	"github.com/tetratelabs/getenvoy/pkg/types"
)

const (
	referenceEnv = "ENVOY_REFERENCE"
)

var (
	// manifestURL defines location of the GetEnvoy manifest.
	manifestURL = &url.URL{
		Scheme: "https",
		Host:   "tetrate.bintray.com",
		Path:   "/getenvoy/manifest.json",
	}
)

// GetURL returns location of the GetEnvoy manifest.
func GetURL() string {
	return manifestURL.String()
}

// SetURL sets location of the GetEnvoy manifest.
func SetURL(rawurl string) error {
	otherURL, err := url.Parse(rawurl)
	if err != nil || otherURL.Host == "" || otherURL.Scheme == "" {
		return errors.Errorf("%q is not a valid manifest URL", rawurl)
	}
	manifestURL = otherURL
	return nil
}

// NewKey creates a manifest key based on the reference it is given
func NewKey(reference string) (*Key, error) {
	// This enables us to parameterize Docker images
	if reference == "@" {
		reference = os.Getenv(referenceEnv)
	}
	ref, err := types.ParseReference(reference)
	if err != nil {
		return nil, err
	}
	key := Key{Flavor: ref.Flavor, Version: ref.Version, Platform: ref.Platform}
	// If platform is empty, fill it in.
	if key.Platform == "" {
		key.Platform = platform()
	}
	return &key, nil
}

// Key is the primary key used to locate Envoy builds in the manifest
type Key types.Reference

func (k Key) String() string {
	return fmt.Sprintf("%v:%v/%v", k.Flavor, k.Version, k.Platform)
}

// Locate returns the location of the binary for the passed parameters from the passed manifest
// The build version is searched for as a prefix of the OperatingSystemVersion.
// If the OperatingSystemVersion is empty it returns the first build listed for that operating system
func Locate(key *Key) (string, error) {
	if key == nil {
		return "", errors.New("passed key was nil")
	}
	log.Debugf("retrieving manifest %s", GetURL())
	manifest, err := fetch(GetURL())
	if err != nil {
		return "", err
	}
	return LocateBuild(key, manifest)
}

// LocateBuild returns the downloadLocationURL of the associated envoy binary in the manifest using the input key
func LocateBuild(key *Key, manifest *api.Manifest) (string, error) {
	// This is pretty horrible... Not sure there is a nicer way though.
	if manifest.Flavors[key.Flavor] != nil && manifest.Flavors[key.Flavor].Versions[key.Version] != nil {
		for _, build := range manifest.Flavors[key.Flavor].Versions[key.Version].Builds {
			normalizedPlatform := types.PlatformFromEnum(build.Platform.String())
			if normalizedPlatform == key.Platform {
				return build.DownloadLocationUrl, nil
			}
		}
	}
	return "", fmt.Errorf("unable to find matching GetEnvoy build for reference %q", key)
}
