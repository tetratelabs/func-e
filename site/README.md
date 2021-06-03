# Envoy site

The following files are hosted on https://getenvoy.io, specifically via [Netlify redirects](https://github.com/tetratelabs/getenvoy.io/blob/master/site/static/_redirects):

Latest master merge:
* https://getenvoy.io/install.sh -> [./install.sh](install.sh)
* https://getenvoy.io/envoy-versions.json -> [./envoy-versions.json](envoy-versions.json)
  * this is verified by [envoy-versions_test.go](envoy-versions_test.go)
* https://getenvoy.io/envoy-versions-schema.json -> [./envoy-versions-schema.json](envoy-versions-schema.json)
* https://www.getenvoy.io/usage/ -> [./usage.md](usage.md)
  * this is verified by [usage_md_test.go](../internal/cmd/usage_md_test.go)
