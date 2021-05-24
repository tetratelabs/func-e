# NAME

getenvoy - Download and run Envoy

# SYNOPSIS

getenvoy

```
[--envoy-versions-url]=[value]
[--help|-h]
[--home-dir]=[value]
[--version|-v]
```

**Usage**:

```
getenvoy [GLOBAL OPTIONS] command [COMMAND OPTIONS] [ARGUMENTS...]
```

# GLOBAL OPTIONS

**--envoy-versions-url**="": URL of Envoy versions JSON

**--help, -h**: show help

**--home-dir**="": GetEnvoy home directory (location of downloaded versions and run archives)

**--version, -v**: print the version


# COMMANDS

## run

Run Envoy as <version> with <args> as arguments, collecting process state on termination

## versions

List available Envoy versions

## install

Download and install a <version> of Envoy

