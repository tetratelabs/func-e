module github.com/tetratelabs/getenvoy

go 1.13

require (
	bitbucket.org/creachadair/shell v0.0.6
	github.com/asaskevich/govalidator v0.0.0-20200907205600-7a23bdc65eef
	github.com/containerd/containerd v1.4.1
	github.com/deislabs/oras v0.8.1
	github.com/docker/distribution v2.7.1+incompatible
	github.com/envoyproxy/go-control-plane v0.9.8-0.20201019204000-12785f608982
	github.com/ghodss/yaml v1.0.0
	github.com/golang/protobuf v1.4.3
	github.com/manifoldco/promptui v0.0.0-00010101000000-000000000000
	github.com/mattn/go-isatty v0.0.12
	github.com/mattn/go-shellwords v1.0.10
	github.com/mholt/archiver v3.1.1+incompatible
	github.com/mitchellh/go-homedir v1.1.0
	github.com/onsi/ginkgo v1.14.2
	github.com/onsi/gomega v1.10.3
	github.com/opencontainers/image-spec v1.0.1
	github.com/otiai10/copy v1.2.0
	github.com/pkg/errors v0.9.1
	github.com/rakyll/statik v0.0.0-00010101000000-000000000000
	github.com/schollz/progressbar/v2 v2.15.0
	github.com/shirou/gopsutil v3.20.10+incompatible
	github.com/shurcooL/httpfs v0.0.0-20190707220628-8d4bc4ba7749
	github.com/spf13/cobra v1.1.1
	github.com/stretchr/testify v1.6.1
	github.com/tetratelabs/getenvoy-package v0.4.0
	github.com/tetratelabs/log v0.0.0-20190710134534-eb04d1e84fb8
	github.com/tetratelabs/multierror v1.1.0
	gotest.tools v2.2.0+incompatible
	istio.io/api v0.0.0-20201120175956-c2df7c41fd8e
	istio.io/istio v0.0.0-20201123050314-d5abe3ea1b99
	rsc.io/letsencrypt v0.0.3 // indirect
)

replace github.com/manifoldco/promptui => github.com/yskopets/promptui v0.7.1-0.20200429230902-361491009c11

replace github.com/rakyll/statik => github.com/yskopets/statik v0.1.8-0.20200501213002-c2d8dcc79889
