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
