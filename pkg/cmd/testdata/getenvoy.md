% getenvoy 8

# NAME

getenvoy - Manage Envoy lifecycle including fetching binaries and collection of process state.

# SYNOPSIS

getenvoy

```
[--help|-h]
[--home-dir]=[value]
[--version|-v]
```

**Usage**:

```
getenvoy [GLOBAL OPTIONS] command [COMMAND OPTIONS] [ARGUMENTS...]
```

# GLOBAL OPTIONS

**--help, -h**: show help

**--home-dir**="": GetEnvoy home directory (location of downloaded artifacts, caches, etc)

**--version, -v**: print the version


# COMMANDS

## run

Runs Envoy and collects process state on exit. Available builds can be retrieved using `getenvoy list`.

## list

List available Envoy version references you can run

## fetch

Downloads a version of Envoy. Available builds can be retrieved using `getenvoy list`.

