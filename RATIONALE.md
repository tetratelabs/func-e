# Notable rationale of func-e

## Why do we use Travis and GitHub Actions instead of only GitHub Actions?
We use Travis to run CentOS integration tests on arm64 until [GitHub Actions supports it](https://github.com/actions/virtual-environments/issues/2552).

This is an alternative to using emulation instead. Using emulation (ex via `setup-qemu-action` and
`setup-buildx-action`) re-introduces problems we eliminated with Travis. For example, not only would
runners take longer to execute (as emulation is slower than native arch), but there is more setup,
and that setup executes on every change. This setup takes time and uses rate-limited registries. It
also introduces potential for compatibility issues when we move off Docker due to its recent
[licensing changes](https://www.docker.com/pricing).

It is true that Travis has a different syntax and that also could fail. However, the risk of
failure is low. What we gain from running end-to-end or packaging tests on Travis is knowledge that
we broke our [Makefile](Makefile) or that there's an unknown new dependency of Envoy® (such as a
change to the floor version of glibc). While these are unlikely to occur, running tests are still
important.

The only place we use emulation is [publishing internal images](.github/workflows/internal-images.yml).
This is done for convenience and occurs once per day, so duration and rate limiting effects are
not important. If something breaks here, the existing images would become stale, so it isn't as
much an emergency as if we put emulation in the critical path (ex in PR tests.)

At the point when GitHub Actions supports free arm64 runners, we can simplify by removing Travis.
Azure DevOps Pipelines already supports arm64, so it is possible GitHub Actions will in the future.

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

---
[xdg]: https://specifications.freedesktop.org/basedir-spec/basedir-spec-latest.html
