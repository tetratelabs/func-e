+++
title = "getenvoy run"
type = "reference"
parent = "getenvoy"
command = "run"
+++
## getenvoy run

Runs Envoy and collects process state on exit. Available builds can be retrieved using `getenvoy list`.

```
getenvoy run reference [flags] [-- <envoy-args>]
```

### Examples

```
# Run using a manifest reference.
getenvoy run ENVOY_VERSION -- --config-path ./bootstrap.yaml

# List available Envoy flags.
getenvoy run ENVOY_VERSION -- --help

```

### Options

```
  -h, --help   help for run
```

### Options inherited from parent commands

```
      --home-dir string   GetEnvoy home directory (location of downloaded artifacts, caches, etc)
```

### SEE ALSO

* [getenvoy](/reference/getenvoy)	 - Fetch and run Envoy

