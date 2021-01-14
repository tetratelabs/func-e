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
	"github.com/deislabs/oras/pkg/oras"
	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
)

var (
	pushOpts = []oras.PushOpt{
		oras.WithConfigMediaType(ConfigMediaType),
		oras.WithNameValidation(nil),
	}
)

// PusherOpts represents options for Pusher
type PusherOpts struct {
	AllowInsecure bool
	UseHTTP       bool
}

// NewPusherOpts returns a default PusherOpts instance
func NewPusherOpts() PusherOpts {
	return PusherOpts{
		AllowInsecure: false,
		UseHTTP:       false,
	}
}

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
				// nolint:gosec this option is only enabled when the user specify the insecure flag.
				InsecureSkipVerify: true,
			},
		}
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
	return &Pusher{resolver: resolver}, nil
}

// Push pushes the image to the registry
func (p *Pusher) Push(imagePath string, imageRef string) (ocispec.Descriptor, error) {
	ctx := context.Background()

	image, err := newWasmImage(imageRef, imagePath)
	if err != nil {
		return ocispec.Descriptor{}, fmt.Errorf("push failed: %w", err)
	}

	manifest, err := oras.Push(ctx, p.resolver, image.ref, image.store, image.layers, pushOpts...)
	if err != nil {
		return manifest, fmt.Errorf("push failed: %w", err)
	}

	return manifest, nil
}
