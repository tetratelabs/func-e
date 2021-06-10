[![Build](https://github.com/tetratelabs/getenvoy/workflows/build/badge.svg)](https://github.com/tetratelabs/getenvoy)
[![Coverage](https://codecov.io/gh/tetratelabs/getenvoy/branch/master/graph/badge.svg)](https://codecov.io/gh/tetratelabs/getenvoy)
[![Go Report Card](https://goreportcard.com/badge/github.com/tetratelabs/getenvoy)](https://goreportcard.com/report/github.com/tetratelabs/getenvoy)
[![License](https://img.shields.io/badge/license-Apache%202.0-blue.svg)](LICENSE)

# GetEnvoy

GetEnvoy makes running [Envoy®](https://www.envoyproxy.io/) easy.

The quickest way to try the command-line interface is an in-lined configuration.
```bash
# Download the latest release as /usr/local/bin/getenvoy https://github.com/tetratelabs/getenvoy/releases
$ curl -L https://getenvoy.io/install.sh | bash -s -- -b /usr/local/bin
# Run the admin server on http://localhost:9901
$ getenvoy run --config-yaml "admin: {address: {socket_address: {address: '127.0.0.1', port_value: 9901}}}"
```

Run `getenvoy help` or have a look at the [Usage Docs](site/usage.md) for more information.

-----
Envoy® is a registered trademark of The Linux Foundation in the United States and/or other countries
