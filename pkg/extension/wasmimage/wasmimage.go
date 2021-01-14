// Copyright 2021 Tetrate
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

package wasmimage

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	orascnt "github.com/deislabs/oras/pkg/content"
	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
)

// wasmImage represents an OCI-compliant wasm image
type wasmImage struct {
	ref      string
	name     string
	contents []byte

	store  *orascnt.Memorystore
	layers []ocispec.Descriptor
}

// newWasmImage returns a new wasmImage instance
func newWasmImage(ref string, path string) (*wasmImage, error) {
	if err := validateFile(path); err != nil {
		return nil, fmt.Errorf("invalid wasm binary: %w", err)
	}

	name := filepath.Base(path)

	contents, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read %#v: %w", path, err)
	}

	store := orascnt.NewMemoryStore()
	contentLayerDescriptor := store.Add(name, ContentLayerMediaType, contents)

	layers := []ocispec.Descriptor{
		contentLayerDescriptor,
	}

	return &wasmImage{
		ref:      ref,
		name:     name,
		contents: contents,
		layers:   layers,
		store:    store,
	}, nil
}

func validateFile(path string) error {
	if _, err := os.Stat(path); err != nil {
		return fmt.Errorf("%#v does not exist: %w", path, err)
	}

	if ext := filepath.Ext(path); ext != ".wasm" {
		return fmt.Errorf("%#v is not a wasm binary", path)
	}

	return nil
}
