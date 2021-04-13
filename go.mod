module github.com/tetratelabs/getenvoy

// This project uses go:embed, so requires minimally go 1.16
go 1.16

require (
	bitbucket.org/creachadair/shell v0.0.6
	github.com/Masterminds/semver/v3 v3.1.1
	github.com/Microsoft/go-winio v0.4.17-0.20210211115548-6eac466e5fa3 // indirect
	github.com/StackExchange/wmi v0.0.0-20210224194228-fe8f1750fd46 // indirect
	github.com/andybalholm/brotli v1.0.1 // indirect
	github.com/asaskevich/govalidator v0.0.0-20210307081110-f21760c49a8d
	github.com/containerd/cgroups v0.0.0-20210114181951-8a68de567b68 // indirect
	github.com/containerd/containerd v1.4.4
	github.com/containerd/continuity v0.0.0-20210208174643-50096c924a4e // indirect
	github.com/deislabs/oras v0.11.1
	github.com/docker/distribution v2.7.1+incompatible
	github.com/docker/docker v20.10.5+incompatible // indirect
	github.com/docker/go-metrics v0.0.1 // indirect
	github.com/envoyproxy/go-control-plane v0.9.8-0.20201019204000-12785f608982
	github.com/envoyproxy/protoc-gen-validate v0.5.1 // indirect
	github.com/go-ole/go-ole v1.2.5 // indirect
	github.com/gogo/protobuf v1.3.2 // indirect
	github.com/golang/groupcache v0.0.0-20210331224755-41bb18bfe9da // indirect
	github.com/golang/snappy v0.0.3 // indirect
	github.com/hashicorp/errwrap v1.1.0 // indirect
	github.com/imdario/mergo v0.3.11 // indirect
	github.com/klauspost/compress v1.11.13 // indirect
	github.com/klauspost/pgzip v1.2.5 // indirect
	github.com/lunixbochs/vtclean v1.0.0 // indirect
	github.com/manifoldco/promptui v0.8.0
	github.com/mattn/go-colorable v0.1.8 // indirect
	github.com/mattn/go-isatty v0.0.12
	github.com/mattn/go-runewidth v0.0.12 // indirect
	github.com/mattn/go-shellwords v1.0.11
	github.com/mholt/archiver/v3 v3.5.0
	github.com/mitchellh/go-homedir v1.1.0
	github.com/moby/term v0.0.0-20201216013528-df9cb8a40635 // indirect
	github.com/opencontainers/image-spec v1.0.1
	github.com/openshift/api v3.9.1-0.20191008181517-e4fd21196097+incompatible // indirect
	github.com/otiai10/copy v1.5.1
	github.com/pierrec/lz4/v4 v4.1.4 // indirect
	github.com/pkg/errors v0.9.1
	github.com/prometheus/client_golang v1.10.0 // indirect
	github.com/prometheus/common v0.20.0 // indirect
	github.com/russross/blackfriday/v2 v2.1.0 // indirect
	github.com/schollz/progressbar/v3 v3.7.6
	github.com/shirou/gopsutil/v3 v3.21.3
	github.com/spf13/cobra v1.1.3
	github.com/stretchr/testify v1.7.0
	github.com/tetratelabs/log v0.0.0-20210323000454-90a3a3e141b5
	github.com/tetratelabs/multierror v1.1.0
	// Match data/extension/init/templates/tinygo/*/default/go.mod_ See RATIONALE.md for why
	github.com/tetratelabs/proxy-wasm-go-sdk v0.1.1
	github.com/tklauser/go-sysconf v0.3.5 // indirect
	github.com/ulikunitz/xz v0.5.10 // indirect
	go.opencensus.io v0.23.0 // indirect
	golang.org/x/crypto v0.0.0-20210322153248-0c34fe9e7dc2 // indirect
	golang.org/x/net v0.0.0-20210405180319-a5a99cb37ef4 // indirect
	golang.org/x/sync v0.0.0-20210220032951-036812b2e83c // indirect
	golang.org/x/sys v0.0.0-20210403161142-5e06dd20ab57 // indirect
	golang.org/x/term v0.0.0-20210406210042-72f3dc4e9b72 // indirect
	google.golang.org/genproto v0.0.0-20210406143921-e86de6bf7a46 // indirect
	google.golang.org/protobuf v1.26.0
	gopkg.in/yaml.v3 v3.0.0-20210107192922-496545a6307b
	istio.io/api v0.0.0-20210205171209-1744f93fc684
	istio.io/istio v0.0.0-20210308183630-97e10d79b8b5
	k8s.io/apiextensions-apiserver v0.19.9 // indirect
	k8s.io/klog/v2 v2.4.0 // indirect
	k8s.io/utils v0.0.0-20201110183641-67b214c5f920 // indirect
	rsc.io/letsencrypt v0.0.3 // indirect
	sigs.k8s.io/yaml v1.2.0
)

// Resolve import problems caused by using istio, currently istio/istio@1.8.4
// See https://github.com/istio/istio/blob/1.8.4/go.mod and go.sum
replace (
	// istio 1.8.4 deps
	github.com/envoyproxy/go-control-plane => github.com/envoyproxy/go-control-plane v0.9.8-0.20201019204000-12785f608982

	google.golang.org/grpc => google.golang.org/grpc v1.33.3
	istio.io/api => istio.io/api v0.0.0-20210205171209-1744f93fc684
	istio.io/istio => istio.io/istio v0.0.0-20210308183630-97e10d79b8b5

	// Latest patch of k8s version included in istio 1.8.4
	k8s.io/api => k8s.io/api v0.19.9
	k8s.io/apiextensions-apiserver => k8s.io/apiextensions-apiserver v0.19.9
	k8s.io/apimachinery => k8s.io/apimachinery v0.19.9
	k8s.io/client-go => k8s.io/client-go v0.19.9
)

// Handle ambiguous import due istio imports
exclude (
	github.com/Azure/go-autorest v10.8.1+incompatible
	github.com/hashicorp/consul v1.3.1
)
