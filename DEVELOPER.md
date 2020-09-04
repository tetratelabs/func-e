# Developer Documentation

## How To

### How to Build

#### getenvoy binary

Run:
```shell
make build
```
which will produce a binary at `./build/bin/$(go env GOOS)/$(go env GOARCH)/getenvoy`.

#### Docker build images for Wasm extensions

Run:
```shell
make builders
```
which will produce `Docker` build images, e.g.
* `getenvoy/extension-rust-builder:latest`

### How to run Unit Tests

Run:
```shell
make test
```

### How to collect Test Coverage

```shell
make coverage
```

### How to create and build a Wasm extension

To create a new Wasm extension, run:
```shell
getenvoy extension init my-extension
```
and follow the wizard.

> NOTE: At the moment, `Rust` extensions have a dependency on a private `GitHub` repository [tetratelabs/envoy-wasm-rust-sdk](https://github.com/tetratelabs/envoy-wasm-rust-sdk).
>
> In practice, it means that `Rust` toolchain (`cargo`) will have to pass through [GitHub authenticatation]() to be able to fetch the source code of [tetratelabs/envoy-wasm-rust-sdk](https://github.com/tetratelabs/envoy-wasm-rust-sdk).
>
> For more details see a section on [SSH authentication](https://doc.rust-lang.org/cargo/appendix/git-authentication.html#ssh-authentication) in the [Cargo Book](https://doc.rust-lang.org/cargo/).

To build a Wasm extension on `Mac OS`, do the following:
1. [Configure SSH agent](https://help.github.com/en/github/authenticating-to-github/generating-a-new-ssh-key-and-adding-it-to-the-ssh-agent#adding-your-ssh-key-to-the-ssh-agent)
2. Run:
   ```shell
   cd my-new-extension
   
   getenvoy extension build --toolchain-container-options \
     '--mount type=bind,src=/run/host-services/ssh-auth.sock,target=/run/host-services/ssh-auth.sock -e SSH_AUTH_SOCK=/run/host-services/ssh-auth.sock'
   ```

### How to run e2e Tests

Run:
```shell
make e2e
```

> NOTE: At the moment, `Rust` extensions have a dependency on a private `GitHub` repository [tetratelabs/envoy-wasm-rust-sdk](https://github.com/tetratelabs/envoy-wasm-rust-sdk).
>
> In practice, it means that `Rust` toolchain (`cargo`) will have to pass through [GitHub authenticatation]() to be able to fetch the source code of [tetratelabs/envoy-wasm-rust-sdk](https://github.com/tetratelabs/envoy-wasm-rust-sdk).
>
> For more details see a section on [SSH authentication](https://doc.rust-lang.org/cargo/appendix/git-authentication.html#ssh-authentication) in the [Cargo Book](https://doc.rust-lang.org/cargo/).

To run e2e tests on `Mac OS`, do the following:
1. [Configure SSH agent](https://help.github.com/en/github/authenticating-to-github/generating-a-new-ssh-key-and-adding-it-to-the-ssh-agent#adding-your-ssh-key-to-the-ssh-agent)
2. Run:
   ```shell
   bash -c '
   set -e

   #
   # restore original ownership over the SSH agent socket (mounted inside container)
   #
   trap "docker run --rm -t --mount type=bind,src=/run/host-services/ssh-auth.sock,target=/run/host-services/ssh-auth.sock busybox chown 0 /run/host-services/ssh-auth.sock" EXIT

   #
   # pass ownership over the SSH agent socket (mounted inside container) to the current user
   #
   docker run --rm -t --mount type=bind,src=/run/host-services/ssh-auth.sock,target=/run/host-services/ssh-auth.sock busybox chown $(id -u) /run/host-services/ssh-auth.sock

   #
   # Run e2e tests in the following context:
   #  1. Pass SSH agent socket (so that build containers could download private dependencies)
   #  2. Override location of Cargo cache (so that all extensions and their build containers could share the same cache)
   #
   E2E_BUILTIN_TOOLCHAIN_CONTAINER_OPTIONS=" \
   --mount type=bind,src=/run/host-services/ssh-auth.sock,target=/run/host-services/ssh-auth.sock \
   -e SSH_AUTH_SOCK=/run/host-services/ssh-auth.sock \
   -v /tmp/cache/getenvoy:/tmp/cache/getenvoy \
   -e CARGO_HOME=/tmp/cache/getenvoy/extension/rust-builder/cargo" \
   make e2e
   '
   ```
