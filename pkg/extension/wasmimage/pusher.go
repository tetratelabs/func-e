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
func NewPusher(insecure bool, useHTTP bool) (*Pusher, error) {
	client := http.DefaultClient
	if insecure {
		client.Transport = &http.Transport{
			TLSClientConfig: &tls.Config{
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
func (p *Pusher) Push(image *WasmImage) (ocispec.Descriptor, error) {
	ctx := context.Background()

	manifest, err := oras.Push(ctx, p.resolver, image.ref, image.store, image.layers, pushOpts...)
	if err != nil {
		return manifest, fmt.Errorf("push failed: %w", err)
	}

	return manifest, nil
}
