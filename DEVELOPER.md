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

Run:
```shell
make e2e
```
