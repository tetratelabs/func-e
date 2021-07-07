[![Build](https://github.com/tetratelabs/func-e/workflows/build/badge.svg)](https://github.com/tetratelabs/func-e)
[![Coverage](https://codecov.io/gh/tetratelabs/func-e/branch/master/graph/badge.svg)](https://codecov.io/gh/tetratelabs/func-e)
[![Go Report Card](https://goreportcard.com/badge/github.com/tetratelabs/func-e)](https://goreportcard.com/report/github.com/tetratelabs/func-e)
[![License](https://img.shields.io/badge/license-Apache%202.0-blue.svg)](LICENSE)

# func-e

func-e (pronounced funky) makes running [Envoy®](https://www.envoyproxy.io/) easy.

The quickest way to try the command-line interface is an in-lined configuration.
```bash
# Download the latest release as /usr/local/bin/func-e https://github.com/tetratelabs/func-e/releases
$ curl https://func-e.io/install.sh | bash -s -- -b /usr/local/bin
# Run the admin server on http://localhost:9901
$ func-e run --config-yaml "admin: {address: {socket_address: {address: '127.0.0.1', port_value: 9901}}}"
```

Run `func-e help` or have a look at the [Usage Docs](USAGE.md) for more information.

-----
Envoy® is a registered trademark of The Linux Foundation in the United States and/or other countries
