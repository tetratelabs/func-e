# GetEnvoy End-to-end (e2e) tests

This directory holds the end-to-end tests for `getenvoy`.

By default, end-to-end (e2e) tests verify a `getenvoy` binary built from [/pkg][../../pkg]

Ex run this from the project root:
```shell
make e2e
```

You can override the binary tested by setting `E2E_GETENVOY_BINARY` to an alternative location, for example a release.

## Version of Envoy under test
The envoy version used in tests default to what's in [/pkg/reference.txt](../../pkg/reference.txt).

## Development Notes

### Don't share code between here and /pkg
This is an end-to-end test of the `getenvoy` binary: it is easy to get confused about what is happening when some code
is in the binary and other shared. Please refrain from using utilities in [/pkg](../../pkg). Place e2e utilities here
instead.

We historically added functions into main only for e2e and left them after they became unused. Adding code into /pkg
also effects its code health statistics. Treat e2e as a separate project even though it shares a [go.mod](../../go.mod)
with /pkg.

### Be careful when adding dependencies
Currently, e2e shares [go.mod](../../go.mod) with [/pkg](../../pkg). This is for simplification in build config and also
details such as linters. However, we carry a risk of dependencies used here ending up accidentally used in /pkg. The IDE
will think this is the same project as /pkg and suggest libraries with auto-complete.

For example, if /pkg used "archiver/v3" accidentally, it would bloat the binary by almost 3MB. For this reason, please
be careful and only add dependencies absolutely needed.

If go.mod ever supports test-only scope, this risk would go away, because IDEs could hide test dependencies from main
auto-complete suggestions. However, it is unlikely go will allow a test scope: https://github.com/golang/go/issues/26913
