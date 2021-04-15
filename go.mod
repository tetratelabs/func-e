module github.com/tetratelabs/getenvoy

// This project uses go:embed, so requires minimally go 1.16
go 1.16

require (
	bitbucket.org/creachadair/shell v0.0.6
	github.com/Masterminds/semver/v3 v3.1.1
	github.com/asaskevich/govalidator v0.0.0-20210307081110-f21760c49a8d
	github.com/containerd/containerd v1.4.4
	github.com/deislabs/oras v0.11.1
	github.com/envoyproxy/go-control-plane v0.9.9-0.20201210154907-fd9021fe5dad
	github.com/manifoldco/promptui v0.8.0
	github.com/mattn/go-isatty v0.0.12
	github.com/mattn/go-shellwords v1.0.11
	github.com/mholt/archiver/v3 v3.5.0
	github.com/mitchellh/go-homedir v1.1.0
	github.com/opencontainers/image-spec v1.0.1
	github.com/otiai10/copy v1.5.1
	github.com/pkg/errors v0.9.1
	github.com/schollz/progressbar/v3 v3.7.6
	github.com/shirou/gopsutil/v3 v3.21.3
	github.com/spf13/cobra v1.1.3
	github.com/stretchr/testify v1.7.0
	github.com/tetratelabs/log v0.0.0-20210323000454-90a3a3e141b5
	github.com/tetratelabs/multierror v1.1.0
	// Match data/extension/init/templates/tinygo/*/default/go.mod_ See RATIONALE.md for why
	github.com/tetratelabs/proxy-wasm-go-sdk v0.1.1
	google.golang.org/protobuf v1.26.0
	gopkg.in/yaml.v3 v3.0.0-20200313102051-9f266ea9e77c
	sigs.k8s.io/yaml v1.2.0
)
