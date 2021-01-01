package wasmimage

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	orascnt "github.com/deislabs/oras/pkg/content"
	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
)

// WasmImage represents an OCI-compliant wasm image
type WasmImage struct {
	ref      string
	name     string
	contents []byte

	store  *orascnt.Memorystore
	layers []ocispec.Descriptor
}

// NewWasmImage returns a new WasmImage instance
func NewWasmImage(ref string, path string) (*WasmImage, error) {
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

	return &WasmImage{
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
