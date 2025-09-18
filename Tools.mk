# This is an alternative to `go tool` for managing Go tools via Makefile.
#
# * This does not pollute the `go.mod` file with tool dependencies.
# * This allows explicit versioning of tools where `go tool` is indirect.
# * This does not integrate as cleanly as `go tool`, and only works with make
# * This does not provide for a go.sum which would lock the versions of tools.

golangci_lint := github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.4.0
gofumpt       := mvdan.cc/gofumpt@v0.9.1
gosimports    := github.com/rinchsan/gosimports/cmd/gosimports@v0.3.8
# sync this with netlify.toml!
hugo          := github.com/gohugoio/hugo@v0.148.1
nwa           := github.com/B1NARY-GR0UP/nwa@v0.7.5
nfpm          := github.com/goreleaser/nfpm/v2/cmd/nfpm@v2.43.1
