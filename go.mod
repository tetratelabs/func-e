module github.com/tetratelabs/getenvoy

go 1.12

require (
	github.com/golang/protobuf v1.3.2
	github.com/spf13/cobra v0.0.5
	github.com/stretchr/testify v1.2.2
	github.com/tetratelabs/getenvoy-package v0.0.0-20190718134531-9487f25b3273
)

replace github.com/tetratelabs/getenvoy-package => ../getenvoy-package
