# Developer Documentation

## How To

### How to Build

#### func-e binary

Run:
```shell
make build
```
which will produce a binary at `./build/bin/$(go env GOOS)/$(go env GOARCH)/func-e`.

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

### How to test the website

Run below, then view with http://localhost:1313/
```shell
make site
```