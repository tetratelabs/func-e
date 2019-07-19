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
	"net/http"

	"github.com/golang/protobuf/jsonpb"
	"github.com/tetratelabs/getenvoy-package/api"
)

// Fetch retrieves and parses a manifest from the URL passed
func Fetch(manifestURL string) (*api.Manifest, error) {
	// #nosec => This is by design, users can call out to wherever they like!
	resp, err := http.Get(manifestURL)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("received %v response code from %v", resp.StatusCode, manifestURL)
	}
	defer func() { _ = resp.Body.Close() }()
	result := api.Manifest{}
	if err := jsonpb.Unmarshal(resp.Body, &result); err != nil {
		return nil, fmt.Errorf("error unmarshalling manifest: %v", err)
	}
	return &result, nil
}
