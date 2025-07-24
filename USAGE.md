# func-e Overview
To run Envoy, execute `func-e run -c your_envoy_config.yaml`. This
downloads and installs the latest version of Envoy for you.

To list versions of Envoy you can use, execute `func-e versions -a`. To
choose one, invoke `func-e use 1.35.0`. This installs into
`$FUNC_E_HOME/versions/1.35.0`, if not already present. You may also use
minor version, such as `func-e use 1.35`.

You may want to override `$ENVOY_VERSIONS_URL` to supply custom builds or
otherwise control the source of Envoy binaries. When overriding, validate
your JSON first: https://archive.tetratelabs.io/release-versions-schema.json

Advanced:
`FUNC_E_PLATFORM` overrides the host OS and architecture of Envoy binaries.
This is used when emulating another platform, e.g. x86 on Apple Silicon M1.
Note: Changing the OS value can cause problems as Envoy has dependencies,
such as glibc. This value must be constant within a `$FUNC_E_HOME`.

# Commands

| Name | Usage |
| ---- | ----- |
| help | Shows how to use a [command] |
| run | Run Envoy with the given [arguments...] until interrupted |
| versions | List Envoy versions |
| use | Sets the current [version] used by the "run" command |
| which | Prints the path to the Envoy binary used by the "run" command |
| --version, -v | Print the version of func-e |

# Environment Variables

| Name | Usage | Default |
| ---- | ----- | ------- |
| FUNC_E_HOME | func-e home directory (location of installed versions and run archives) | ${HOME}/.func-e |
| ENVOY_VERSIONS_URL | URL of Envoy versions JSON | https://archive.tetratelabs.io/envoy/envoy-versions.json |
| FUNC_E_PLATFORM | the host OS and architecture of Envoy binaries. Ex. darwin/arm64 | $GOOS/$GOARCH |
