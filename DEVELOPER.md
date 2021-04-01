# Developer Documentation

## How To

### How to Build

#### getenvoy binary

Run:
```shell
make build
```
which will produce a binary at `./build/bin/$(go env GOOS)/$(go env GOARCH)/getenvoy`.

#### Docker build images for Wasm extensions

Run:
```shell
make builders
```
which will produce `Docker` build images, e.g.
* `getenvoy/extension-rust-builder:latest`

### How to run Unit Tests

Run:
```shell
make test
```

### How to collect Test Coverage

```shell
make coverage
```

### How to create and build a Wasm extension

To create a new Wasm extension, run:
```shell
getenvoy extension init my-extension
```
and follow the wizard.

### How to run e2e Tests

End-to-end (e2e) tests rely on a `getenvoy` binary that defaults to what was built by `make bin`.

In simplest case, execute `make e2e` to run all tests configured.

To constrain tests to one extension language, such as "tinygo" set `E2E_EXTENSION_LANGUAGE` accordingly.

The below ENV variables to effect the e2e execution. These are defined in [main_test.go](test/e2e/main_test.go).
Environment Variable              | Description
--------------------------------- | ------------------------------------------------------------------------------------
`E2E_GETENVOY_BINARY`             | Overrides `getenvoy` binary. Defaults to `$PWD/build/bin/$GOOS/$GOARCH/getenvoy`
`E2E_TOOLCHAIN_CONTAINER_OPTIONS` | Overrides `--toolchain-container-options` in Docker commands. Defaults to "".
`E2E_EXTENSION_LANGUAGE`          | Overrides `--language` in `getenvoy extension` commands. Defaults to "all".
