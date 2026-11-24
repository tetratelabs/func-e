module github.com/tetratelabs/func-e

// Before updating minor check compatibility with Envoy Gateway.
// See https://github.com/envoyproxy/gateway/blob/main/go.mod#L3
// .0 per library convention; the toolchain directive pins the build version.
go 1.27.0

toolchain go1.27.4

require (
	github.com/quasilyte/go-ruleguard/dsl v0.3.23
	github.com/shirou/gopsutil/v4 v4.26.10
	github.com/stretchr/testify v1.12.1
	github.com/ulikunitz/xz/v2 v2.0.0-dev.4
	github.com/urfave/cli-docs/v3 v3.2.0
	github.com/urfave/cli/v3 v3.14.0
	gopkg.in/yaml.v3 v3.0.1
)

require (
	github.com/cpuguy83/go-md2man/v2 v2.0.7 // indirect
	github.com/ebitengine/purego v0.11.2 // indirect
	github.com/go-ole/go-ole v1.2.6 // indirect
	github.com/kr/pretty v0.3.1 // indirect
	github.com/lufia/plan9stats v0.0.0-20220913051719-115f729f3c8c // indirect
	github.com/power-devops/perfstat v0.0.0-20260805114148-88456608a4f6 // indirect
	github.com/rogpeppe/go-internal v1.14.1 // indirect
	github.com/russross/blackfriday/v2 v2.1.0 // indirect
	github.com/tklauser/go-sysconf v0.4.0 // indirect
	github.com/tklauser/numcpus v0.12.0 // indirect
	github.com/ulikunitz/lz v0.6.11 // indirect
	github.com/yusufpapurcu/wmi v1.2.4 // indirect
	go.yaml.in/yaml/v3 v3.0.5 // indirect
	golang.org/x/sys v0.48.0 // indirect
	gopkg.in/check.v1 v1.0.0-20201130134442-10cb98267c6c // indirect
)
