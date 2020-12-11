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

type Pusher struct {
	resolver remotes.Resolver
}

type PusherOpts struct {
	AllowInsecure bool
	UseHTTP       bool
}

func NewPusherOpts() PusherOpts {
	return PusherOpts{
		AllowInsecure: false,
		UseHTTP:       false,
	}
}

func NewPusher(insecure, useHTTP bool) (*Pusher, error) {
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

func (p *Pusher) Push(image *WasmImage) (ocispec.Descriptor, error) {
	ctx := context.Background()

	pushOpts := []oras.PushOpt{
		oras.WithConfigMediaType(ConfigMediaType),
		oras.WithNameValidation(nil),
	}

	manifest, err := oras.Push(ctx, p.resolver, image.ref, image.store, image.layers, pushOpts...)
	if err != nil {
		return manifest, fmt.Errorf("push failed: %w", err)
	}

	return manifest, nil
}
