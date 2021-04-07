module github.com/tetratelabs/getenvoy

// This project uses go:embed, so requires minimally go 1.16
go 1.16

require (
	bitbucket.org/creachadair/shell v0.0.6
	github.com/Masterminds/semver v1.5.0
	github.com/Nvveen/Gotty v0.0.0-20120604004816-cd527374f1e5 // indirect
	github.com/StackExchange/wmi v0.0.0-20190523213315-cbe66965904d // indirect
	github.com/asaskevich/govalidator v0.0.0-20190424111038-f61b66f89f4a
	github.com/containerd/containerd v1.3.2
	github.com/deislabs/oras v0.8.1
	github.com/docker/distribution v2.7.1+incompatible
	github.com/envoyproxy/go-control-plane v0.9.5
	github.com/ghodss/yaml v1.0.0
	github.com/go-ole/go-ole v1.2.4 // indirect
	github.com/golang/protobuf v1.3.5
	github.com/gotestyourself/gotestyourself v2.2.0+incompatible // indirect
	github.com/manifoldco/promptui v0.8.0
	github.com/mattn/go-isatty v0.0.12
	github.com/mattn/go-shellwords v1.0.10
	github.com/mholt/archiver v3.1.1+incompatible
	github.com/mitchellh/go-homedir v1.1.0
	github.com/onsi/ginkgo v1.12.1
	github.com/onsi/gomega v1.10.0
	github.com/opencontainers/image-spec v1.0.1
	github.com/opencontainers/selinux v1.8.0 // indirect
	github.com/otiai10/copy v1.2.0
	github.com/pkg/errors v0.9.1
	github.com/schollz/progressbar/v2 v2.13.2
	github.com/shirou/gopsutil v0.0.0-20190731134726-d80c43f9c984
	github.com/shurcooL/httpfs v0.0.0-20190707220628-8d4bc4ba7749
	github.com/spf13/cobra v0.0.5
	github.com/stretchr/testify v1.7.0
	github.com/tetratelabs/getenvoy-package v0.0.0-20190730071641-da31aed4333e
	github.com/tetratelabs/log v0.0.0-20190710134534-eb04d1e84fb8
	github.com/tetratelabs/multierror v1.1.0
	// Match data/extension/init/templates/tinygo/*/default/go.mod_ See RATIONALE.md for why
	github.com/tetratelabs/proxy-wasm-go-sdk v0.1.1
	istio.io/api v0.0.0-20200227213531-891bf31f3c32
	istio.io/istio v0.0.0-20200304114959-c3c353285578
	rsc.io/letsencrypt v0.0.3 // indirect
)

replace github.com/Azure/go-autorest/autorest => github.com/Azure/go-autorest/autorest v0.11.15

replace github.com/docker/docker => github.com/docker/docker v17.12.1-ce+incompatible

replace github.com/hashicorp/consul => github.com/hashicorp/consul v1.3.1
