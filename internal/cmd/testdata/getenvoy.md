# NAME

getenvoy - Install and run Envoy

# SYNOPSIS

getenvoy

```
[--envoy-versions-url]=[value]
[--home-dir]=[value]
[--version|-v]
```

**Usage**:

```
getenvoy [GLOBAL OPTIONS] command [COMMAND OPTIONS] [ARGUMENTS...]
```

# GLOBAL OPTIONS

**--envoy-versions-url**="": URL of Envoy versions JSON

**--home-dir**="": GetEnvoy home directory (location of installed versions and run archives)

**--version, -v**: print the version


# COMMANDS

## help

Shows how to use a [command]

## run

Run Envoy with the given [arguments...], collecting process state on termination

## versions

List Envoy versions

**--all, -a**: Show all versions including ones not yet installed

## use

Sets the current [version] used by the "run" command, installing as necessary

