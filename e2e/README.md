# GetEnvoy End-to-end (e2e) tests

This directory holds the end-to-end tests for `getenvoy`.

By default, end-to-end (e2e) tests verify a `getenvoy` binary built from [main.go](../main.go)

Ex run this from the project root:
```shell
make e2e
```

You can override the binary tested by setting `E2E_GETENVOY_BINARY` to an alternative location, for example a release.

## Version of Envoy under test
The envoy version used in tests default to what's in [/internal/reference/latest.txt](../internal/reference/latest.txt).

## Development Notes

### Don't share add code to /internal only used here
This is an end-to-end test of the `getenvoy` binary: it is easy to get confused about what is happening when some code
is in the binary and other shared. To avoid confusion, only use code in [/internal](../internal) on an exception basis.

We historically added functions into main only for e2e and left them after they became unused. Adding code into
/internal also effects main code health statistics. Hence, we treat e2e as a separate project even though it shares a
[go.mod](../go.mod) with /internal. Specifically, we don't add code into /internal which only needs to exist here.

### Be careful when adding dependencies
Currently, e2e shares [go.mod](../go.mod) with [/internal](../internal). This is for simplification in build config and
details such as linters. However, we carry a risk of dependencies used here ending up accidentally used in /internal.
The IDE  will think this is the same project as /internal and suggest libraries with auto-complete.

For example, if /internal used "archiver/v3" accidentally, it would bloat the binary by almost 3MB. For this reason,
please be careful and only add dependencies absolutely needed.

If go.mod ever supports test-only scope, this risk would go away, because IDEs could hide test dependencies from main
auto-complete suggestions. However, it is unlikely go will allow a test scope: https://github.com/golang/go/issues/26913
