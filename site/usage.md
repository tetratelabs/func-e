# GetEnvoy CLI Overview
To run Envoy, execute `getenvoy run -c your_envoy_config.yaml`. This
downloads and installs the latest version of Envoy for you.

To list versions of Envoy you can use, execute `getenvoy versions -a`. To
choose one, invoke `getenvoy use 1.18.3`. This installs into
`$GETENVOY_HOME/versions/1.18.3`, if not already present.

You may want to override `$ENVOY_VERSIONS_URL` to supply custom builds or
otherwise control the source of Envoy binaries. When overriding, validate
your JSON first: https://getenvoy.io/envoy-versions-schema.json

# Commands

| Name | Usage |
| ---- | ----- |
| help | Shows how to use a [command] |
| run | Run Envoy with the given [arguments...], collecting process state on termination |
| versions | List Envoy versions |
| use | Sets the current [version] used by the "run" command, installing as necessary |
| --version, -v | Print the version of GetEnvoy |

# Environment Variables

| Name | Usage | Default |
| ---- | ----- | ------- |
| GETENVOY_HOME | GetEnvoy home directory (location of installed versions and run archives) | ${HOME}/.getenvoy |
| ENVOY_VERSIONS_URL | URL of Envoy versions JSON | https://getenvoy.io/envoy-versions.json |
