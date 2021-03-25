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
	"context"
	"crypto/tls"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/containerd/containerd/remotes"
	"github.com/deislabs/oras/pkg/auth/docker"
	orascnt "github.com/deislabs/oras/pkg/content"
	orasctx "github.com/deislabs/oras/pkg/context"
	"github.com/deislabs/oras/pkg/oras"
	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
)

// Puller knows how to fetch wasm images from OCI-compliant registries.
type Puller struct {
	resolver remotes.Resolver
}

// NewPuller returns a new Puller instance.
func NewPuller(insecure, useHTTP bool) (*Puller, error) {
	client := http.DefaultClient

	client.Transport = &http.Transport{
		TLSClientConfig: &tls.Config{
			// this option is only enabled when the user specify the insecure flag.
			InsecureSkipVerify: insecure, // nolint:gosec
		},
	}

	// TODO(musaprg): separate these instructions into another functions
	auth, err := docker.NewClient()
	if err != nil {
		return nil, err
	}
	resolver, err := auth.Resolver(context.Background(), client, useHTTP)
	if err != nil {
		return nil, err
	}
	return &Puller{resolver: resolver}, nil
}

// Pull fetches the specified image from the registry
func (p *Puller) Pull(imageRef, imagePath string) (ocispec.Descriptor, error) {
	ctx := orasctx.Background()
	store := orascnt.NewMemoryStore()

	pullOpts := []oras.PullOpt{
		oras.WithAllowedMediaType(contentLayerMediaType),
		oras.WithPullEmptyNameAllowed(),
	}

	_, layers, err := oras.Pull(ctx, p.resolver, imageRef, store, pullOpts...)
	if err != nil {
		return ocispec.Descriptor{}, fmt.Errorf("pull failed: %w", err)
	}

	if len(layers) != 1 {
		return ocispec.Descriptor{}, fmt.Errorf("invalid number of image layers")
	}
	manifest, image, _ := store.Get(layers[0])

	if err := ioutil.WriteFile(imagePath, image, 0600); err != nil {
		return manifest, fmt.Errorf("failed to write image: %w", err)
	}

	return manifest, nil
}
