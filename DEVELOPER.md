# Developer Documentation

## How To

### How to Build

Make sure you are running the same version of go as indicated in [go.mod](go.mod).

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

### How to generate release assets

To generate release assets, run the below:
```shell
make dist
```

The contents will be in the 'dist/' folder and include the same files as a
[release](https://github.com/tetratelabs/func-e/releases) would, except
signatures would not be the same as production.

Note: this step requires prerequisites for Windows packaging to work. Look at
[msi.yaml](.github/workflows/msi.yaml) for what's needed per-platform.
