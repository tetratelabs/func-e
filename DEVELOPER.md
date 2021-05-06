# Developer Documentation

## How To

### How to Build

#### getenvoy binary

Run:
```shell
make build
```
which will produce a binary at `./build/bin/$(go env GOOS)/$(go env GOARCH)/getenvoy`.

### How to run Unit Tests

Run:
```shell
make test
```

### How to collect Test Coverage

Run:
```shell
make coverage
```

### How to run e2e Tests

End-to-end (e2e) tests rely on a `getenvoy` binary that defaults to what was built by `make bin`.

Run:
```shell
make coverage
```
