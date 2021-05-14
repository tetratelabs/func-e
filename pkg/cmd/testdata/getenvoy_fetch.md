+++
title = "getenvoy fetch"
type = "reference"
parent = "getenvoy"
command = "fetch"
+++
## getenvoy fetch

Downloads a version of Envoy. Available builds can be retrieved using `getenvoy list`.

```
getenvoy fetch <reference> [flags]
```

### Examples

```
# Fetch using a partial manifest reference to retrieve a build suitable for your operating system.
getenvoy fetch VERSION
		
# Fetch using a full manifest reference to retrieve a specific build for Linux. 
getenvoy fetch VERSION/linux-glibc
```

### Options

```
  -h, --help   help for fetch
```

### Options inherited from parent commands

```
      --home-dir string   GetEnvoy home directory (location of downloaded artifacts, caches, etc)
```

### SEE ALSO

* [getenvoy](/reference/getenvoy)	 - Fetch and run Envoy

