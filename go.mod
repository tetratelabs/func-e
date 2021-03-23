module github.com/tetratelabs/getenvoy

go 1.13

require (
	bitbucket.org/creachadair/shell v0.0.6
	github.com/Masterminds/semver v1.5.0
	github.com/Nvveen/Gotty v0.0.0-20120604004816-cd527374f1e5 // indirect
	github.com/asaskevich/govalidator v0.0.0-20200428143746-21a406dcc535
	github.com/containerd/containerd v1.3.4
	github.com/deislabs/oras v0.8.1
	github.com/docker/distribution v2.7.1+incompatible
	github.com/envoyproxy/go-control-plane v0.9.9-0.20210115003313-31f9241a16e6
	github.com/frankban/quicktest v1.11.3 // indirect
	github.com/ghodss/yaml v1.0.0
	github.com/golang/protobuf v1.4.3
	github.com/golang/snappy v0.0.3 // indirect
	github.com/gotestyourself/gotestyourself v2.2.0+incompatible // indirect
	github.com/manifoldco/promptui v0.8.0
	github.com/mattn/go-isatty v0.0.12
	github.com/mattn/go-shellwords v1.0.10
	github.com/mholt/archiver v3.1.1+incompatible
	github.com/mitchellh/go-homedir v1.1.0
	github.com/onsi/ginkgo v1.14.1
	github.com/onsi/gomega v1.10.5
	github.com/opencontainers/image-spec v1.0.1
	github.com/opencontainers/selinux v1.8.0 // indirect
	github.com/otiai10/copy v1.2.0
	github.com/pierrec/lz4 v2.6.0+incompatible // indirect
	github.com/pkg/errors v0.9.1
	github.com/rakyll/statik v0.0.0-00010101000000-000000000000
	github.com/schollz/progressbar/v2 v2.13.2
	github.com/shirou/gopsutil v0.0.0-20190731134726-d80c43f9c984
	github.com/shurcooL/httpfs v0.0.0-20190707220628-8d4bc4ba7749
	github.com/spf13/cobra v1.1.3
	github.com/stretchr/testify v1.7.0
	github.com/tetratelabs/getenvoy-package v0.0.0-20190730071641-da31aed4333e
	github.com/tetratelabs/log v0.0.0-20210323000454-90a3a3e141b5
	github.com/tetratelabs/multierror v1.1.0
	github.com/ulikunitz/xz v0.5.10 // indirect
	gotest.tools v2.2.0+incompatible
	istio.io/api v0.0.0-20210322145030-ec7ef4cd6eaf
	istio.io/istio v0.0.0-20210323064757-d4476bb31e8b
	rsc.io/letsencrypt v0.0.3 // indirect
)

replace github.com/Azure/go-autorest/autorest => github.com/Azure/go-autorest/autorest v0.11.15

replace github.com/docker/docker => github.com/docker/docker v17.12.1-ce+incompatible

replace github.com/hashicorp/consul/api => github.com/hashicorp/consul/api v1.8.1

// Support -a arg until https://github.com/rakyll/statik/pull/113
replace github.com/rakyll/statik => github.com/yskopets/statik v0.1.8-0.20200501213002-c2d8dcc79889

// Pending https://github.com/kubernetes/kube-openapi/pull/220
replace k8s.io/kube-openapi => github.com/howardjohn/kube-openapi v0.0.0-20210104181841-c0b40d2cb1c8
