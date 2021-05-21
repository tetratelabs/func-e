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

### How to run end-to-end Tests

See [test/e2e](e2e) for how to develop or run end-to-end tests
