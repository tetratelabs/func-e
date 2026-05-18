# Notable rationale of func-e

## Why do we have so many environment variables for file locations?

Recently, func-e is used as a library, which means its default locations need
to support layouts that are non-default. We also want to be able to predict the
log directory when running in Docker instead of using a timestamp in the path.

This results in several config locations like this:

| Environment Variable  | Default Path            | API Option          |
|-----------------------|-------------------------|---------------------|
| `FUNC_E_CONFIG_HOME`  | `~/.config/func-e`      | `api.ConfigHome()`  |
| `FUNC_E_DATA_HOME`    | `~/.local/share/func-e` | `api.DataHome()`    |
| `FUNC_E_STATE_HOME`   | `~/.local/state/func-e` | `api.StateHome()`   |
| `FUNC_E_RUNTIME_DIR`  | `/tmp/func-e-${UID}`    | `api.RuntimeDir()`  |
| `FUNC_E_RUN_ID`       | auto-generated          | `api.RunID()`       |

These are conventional to [XDG][xdg], which makes it easier to explain to
people. Also, XDS conventions are used by Prometheus and block/goose, so will
be familiar to some.

In summary, XDS conventions allow dependents like Envoy AI Gateway to brand
their own directories and co-mingle its configuration and logs with those
of func-e when it runs Envoy (the gateway process). It also allows Docker to
export `FUNC_E_RUN_ID=0` to aid in location of key files.

## Why tools/go.mod?

`go tool` lets us run things like linters and hugo without a platform install.
We keep them in `tools/go.mod` instead of the main `go.mod` because func-e is
also imported as a library; tool dependencies should not leak into that graph.

This replaces the former `go run package@version` process, where versions
lived in Makefile commands instead of a normal checked-in module graph.

The biggest tradeoff from our old process is tools are not isolated from each
other. Hugo, golangci-lint, nfpm, etc. share one dependency graph and can
revlock each other in the future.

Another wrinkle is [Go rejects `-modfile` in workspace mode][go-work-modfile].
To use `-modfile=tools/go.mod`, we have to set `GOWORK=off`. This causes cruft
in the root Makefile.

## Why internal/test/httptest?

Go's `net/http/httptest.NewServer` listens on loopback TCP. Network I/O is
[not durably blocking][synctest-blocking], so goroutines stuck on TCP prevent
a [synctest][synctest-pkg] bubble from becoming idle. The fake clock never
advances, and any `time.After` or `time.Sleep` in the code under test hangs.

Our `httptest.NewServer` replaces the TCP listener with `net.Pipe`, following
the pattern in Go's [TestTLSServerWithoutTLSConn][go-serve] and
[TransportCancelRequestBeforeResponseHeaders][go-transport] tests. Pipe
operations block on channels, which are durably blocking, so synctest can
see when the bubble is idle and advance the clock. Retrofitting
`httptest.NewServer` keeps test practice familiar while avoiding the
[blocking][synctest-blocking] behavior.

We also provide `httptest.HTTPClient`, which runs a handler synchronously in
the caller's goroutine with no I/O at all. Tests who need `*http.Client`, but
don't need a real server can use this instead.

## Why "dev-latest" instead of a flag?

func-e is embedded in CI systems and tools like [Envoy AI Gateway][ai-gw]
that may only allow version overrides, not flags. A version string works
everywhere `ENVOY_VERSION` is accepted, including the Go API.

`dev` installs on demand like all other versions. Once pulled, it stays
put. This keeps CI runs stable and reproducible without network access on
every invocation.

`dev-latest` is the explicit refresh trigger. Without it, a cached `dev`
install would never update. CI pipelines and tools that re-warm caches
need a way to pull the latest build without manual intervention.

---
[ai-gw]: https://github.com/envoyproxy/ai-gateway
[xdg]: https://specifications.freedesktop.org/basedir-spec/basedir-spec-latest.html
[go-work-modfile]: https://go.dev/issue/59996
[synctest-pkg]: https://pkg.go.dev/testing/synctest
[synctest-blocking]: https://github.com/golang/go/blob/master/src/testing/synctest/synctest.go#L86-L93
[go-serve]: https://github.com/golang/go/blob/master/src/net/http/serve_test.go#L1767
[go-transport]: https://github.com/golang/go/blob/master/src/net/http/transport_test.go#L3079
