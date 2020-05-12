# Sample Access Logger

Example of an Envoy Access logger.

See [envoy-wasm-rust-sdk/examples/access-logger](https://github.com/tetratelabs/envoy-wasm-rust-sdk/tree/master/examples/access-logger)
for more details.

## How To

### How to Set up Rust

Recommended way:
```shell
rustup target add wasm32-unknown-unknown
```

Alternative way:
* Follow [instructions](https://forge.rust-lang.org/infra/other-installation-methods.html#standalone-installers) how to use a standalone installer

### How To Build

```shell
cargo build:wasm
```

> NOTE: At the moment, this step might fail due to dependency on a private `GitHub` repository [tetratelabs/envoy-wasm-rust-sdk](https://github.com/tetratelabs/envoy-wasm-rust-sdk).
>
> Depending on your environment, you can run into an error similar to:
>
>     Updating git repository `ssh://git@github.com/tetratelabs/envoy-wasm-rust-sdk.git`
>
>     error: failed to load source for a dependency on `envoy-sdk`
> There are a couple options to fix this:
> 1. Either configure SSH agent (see [SSH authentication](https://doc.rust-lang.org/cargo/appendix/git-authentication.html#ssh-authentication) section in the [Cargo Book](https://doc.rust-lang.org/cargo/))
> 2. or let `Cargo` use native `git` client when fetching [such a dependency](https://doc.rust-lang.org/cargo/appendix/git-authentication.html#git-authentication), e.g.
>
>        CARGO_NET_GIT_FETCH_WITH_CLI=true cargo build:wasm

### How to Run unit tests

```shell
cargo test
```
