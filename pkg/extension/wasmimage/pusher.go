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
	"net/http"

	"github.com/containerd/containerd/remotes"
	"github.com/deislabs/oras/pkg/auth/docker"
	orasctx "github.com/deislabs/oras/pkg/context"
	"github.com/deislabs/oras/pkg/oras"
	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
)

// Pusher knows how to push wasm images to OCI-compliant registries.
type Pusher struct {
	resolver remotes.Resolver
}

// NewPusher returns a new Pusher instance.
func NewPusher(insecure, useHTTP bool) (*Pusher, error) {
	client := http.DefaultClient

	if insecure {
		client.Transport = &http.Transport{
			TLSClientConfig: &tls.Config{
				// this option is only enabled when the user specify the insecure flag.
				InsecureSkipVerify: true, // nolint:gosec
			},
		}
	}

	// TODO(musaprg): separate these instructions into another functions
	auth, err := docker.NewClient()
	if err != nil {
		return nil, fmt.Errorf("failed to create Pusher: %w", err)
	}
	resolver, err := auth.Resolver(context.Background(), client, useHTTP)
	if err != nil {
		return nil, fmt.Errorf("failed to create Pusher: %w", err)
	}
	return &Pusher{resolver: resolver}, nil
}

// Push pushes the image to the registry
func (p *Pusher) Push(imagePath, imageRef string) (manifest ocispec.Descriptor, size int64, err error) {
	ctx := orasctx.Background()

	image, err := newWasmImage(imageRef, imagePath)
	if err != nil {
		return ocispec.Descriptor{}, 0, fmt.Errorf("push failed: %w", err)
	}

	pushOpts := []oras.PushOpt{
		oras.WithConfigMediaType(configMediaType),
		oras.WithNameValidation(nil),
	}

	manifest, err = oras.Push(ctx, p.resolver, image.ref, image.store, image.layers, pushOpts...)
	if err != nil {
		return manifest, 0, fmt.Errorf("push failed: %w", err)
	}

	return manifest, image.layers[0].Size, nil
}
