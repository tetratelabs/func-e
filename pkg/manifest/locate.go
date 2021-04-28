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
	"io"
	"net/http"
	"os"

	"github.com/tetratelabs/log"
	"google.golang.org/protobuf/encoding/protojson"

	"github.com/tetratelabs/getenvoy/api"
	"github.com/tetratelabs/getenvoy/pkg/transport"
	"github.com/tetratelabs/getenvoy/pkg/types"
)

const (
	referenceEnv = "ENVOY_REFERENCE"
)

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

func (k *Key) String() string {
	return fmt.Sprintf("%v:%v/%v", k.Flavor, k.Version, k.Platform)
}

// FetchManifest returns a manifest from a remote URL. eg global.manifestURL.
func FetchManifest(manifestURL string) (*api.Manifest, error) {
	log.Debugf("retrieving manifest %s", manifestURL)
	// #nosec => This is by design, users can call out to wherever they like!
	resp, err := transport.Get(manifestURL)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close() //nolint

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("received %v response code from %v", resp.StatusCode, manifestURL)
	}
	body, err := io.ReadAll(resp.Body) // fully read the response
	if err != nil {
		return nil, fmt.Errorf("error reading %s: %w", manifestURL, err)
	}

	result := api.Manifest{}
	if err := protojson.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("error unmarshalling manifest: %w", err)
	}
	return &result, nil
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
