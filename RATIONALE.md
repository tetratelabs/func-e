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

---
[xdg]: https://specifications.freedesktop.org/basedir-spec/basedir-spec-latest.html
[go-work-modfile]: https://go.dev/issue/59996
