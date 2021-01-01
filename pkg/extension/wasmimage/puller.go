package wasmimage

import (
	"context"
	"crypto/tls"
	"fmt"
	"net/http"

	"github.com/containerd/containerd/remotes"
	"github.com/deislabs/oras/pkg/auth/docker"
	orascnt "github.com/deislabs/oras/pkg/content"
	"github.com/deislabs/oras/pkg/oras"
)

var (
	pullOpts = []oras.PullOpt{
		oras.WithAllowedMediaType(ContentLayerMediaType),
		oras.WithPullEmptyNameAllowed(),
	}
)

// Puller knows how to fetch wasm images from OCI-compliant registries.
type Puller struct {
	resolver remotes.Resolver
}

// NewPuller returns a new Puller instance.
func NewPuller(insecure bool, useHTTP bool) (*Puller, error) {
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
	return &Puller{resolver: resolver}, nil
}

// Pull fetches the specified image from the registry
func (p *Puller) Pull(ctx context.Context, ref string) (*WasmImage, error) {	
	store := orascnt.NewMemoryStore()

	_, layers, err := oras.Pull(ctx, p.resolver, ref, store, pullOpts...)
	if err != nil {
		return nil, fmt.Errorf("pull failed: %w", err)
	}

	if len(layers) != 1 {
		return nil, fmt.Errorf("invalid number of image layers")
	}
	_, image, _ := store.Get(layers[0])

	return &WasmImage{
		ref: ref,
		contents: image,
		store: store,
		layers: layers,
	}, nil
}