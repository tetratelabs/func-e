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
	"bytes"
	"fmt"
	"os"
	"path/filepath"

	orascnt "github.com/deislabs/oras/pkg/content"
	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
)

var (
	magic = []byte{0x00, 0x61, 0x73, 0x6d}
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
func newWasmImage(ref, path string) (*wasmImage, error) {
	if _, err := os.Stat(path); err != nil {
		return nil, fmt.Errorf("%#v does not exist: %w", path, err)
	}

	if !isWasmBinary(path) {
		return nil, fmt.Errorf("invalid wasm binary")
	}

	name := filepath.Base(path)

	contents, err := os.ReadFile(filepath.Clean(path))
	if err != nil {
		return nil, fmt.Errorf("failed to read %#v: %w", path, err)
	}

	store := orascnt.NewMemoryStore()
	contentLayerDescriptor := store.Add(name, contentLayerMediaType, contents)

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

// isWasmBinary checks whether the file is valid wasm binary
func isWasmBinary(path string) bool {
	f, err := os.Open(filepath.Clean(path))
	defer f.Close() //nolint
	if err != nil {
		return false
	}
	buffer := make([]byte, len(magic))
	n, err := f.Read(buffer)
	if n != len(magic) || err != nil {
		return false
	}
	return bytes.Equal(buffer, magic)
}
