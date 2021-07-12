# func-e End-to-end (e2e) tests

This directory holds the end-to-end tests for `func-e`.

By default, end-to-end (e2e) tests verify a `func-e` binary built from [main.go](../main.go)

Ex. in native go commands:
```bash
go build --ldflags '-s -w' .
go test -parallel 1 -v -failfast ./e2e
```

Ex. using `make`
```bash
make e2e
```

Tests look for `func-e` (or `func-e.exe` in Windows), in the current directory. When run via `make`, the binary location
is read from `E2E_FUNC_E_PATH`: the `goreleaser` dist directory of the current platform. Ex. `dist/func-e_darwin_amd64`

If the `func-e` version is a snapshot and "envoy-versions.json" exists, tests run against the local. This allows local
development and pull requests to verify changes not yet [published](https://archive.tetratelabs.io/envoy/envoy-versions.json)
or those that effect the [schema](https://archive.tetratelabs.io/release-versions-schema.json).

## Version of Envoy under test
The envoy version used in tests default to [/internal/version/last_known_envoy.txt](../internal/version/last_known_envoy.txt).

## Development Notes

### Don't share add code to /internal only used here
This is an end-to-end test of the `func-e` binary: it is easy to get confused about what is happening when some code
is in the binary and other shared. To avoid confusion, only use code in [/internal](../internal) on an exception basis.

We historically added functions into main only for e2e and left them after they became unused. Adding code into
/internal also effects main code health statistics. Hence, we treat e2e as a separate project even though it shares a
[go.mod](../go.mod) with /internal. Specifically, we don't add code into /internal which only needs to exist here.

### Be careful when adding dependencies
Currently, e2e shares [go.mod](../go.mod) with [/internal](../internal). This is for simplification in build config and
details such as linters. However, we carry a risk of dependencies used here ending up accidentally used in /internal.
The IDE will think this is the same project as /internal and suggest libraries with auto-complete.

For example, if /internal used "archiver/v3" accidentally, it would bloat the binary by almost 3MB. For this reason,
please be careful and only add dependencies absolutely needed.

If go.mod ever supports test-only scope, this risk would go away, because IDEs could hide test dependencies from main
auto-complete suggestions. However, it is unlikely go will allow a test scope: https://github.com/golang/go/issues/26913
