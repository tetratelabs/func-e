# Notable Rationale

## go:embed for embedding example and init templates

This project resolves templates into working examples or extensions via `getenvoy example add` and
`getenvoy extension init`. Input [example](data/example/init/templates) and [extension](data/extension/init/templates)
templates are embedded in the `getenvoy` binary for user convenience.

We implement embedding with a Go 1.16+ directive `//go:embed templates/*`, which presents an `fs.FS` interface of the
directory tree. Using this requires no extra build steps. Ex. `go build -o getenvoy ./cmd/getenvoy/main.go` works.

However, there are some constraints to understand. It is accepted that while imperfect, this solution is an acceptable
alternative to a custom solution.

### Embedded files must be in a sub-path
The go source file embedding content via `go:embed` must reference paths in the current directory tree. In other words,
paths like `../../templates` are not allowed.

This means we have to move the embedded directory tree where code refers to it, or make an accessor utility for each
directory root. We use the latter pattern, specifically here:
* [data/example/init/templates.go](data/example/init/templates.go) 
* [data/extension/init/templates.go](data/extension/init/templates.go)

See https://pkg.go.dev/embed#hdr-Directives for more information.

### Limitations of `go:embed` impact extension init templates
`getenvoy extension init` creates a workspace directory for a given category and programming language. Some constraints
of `go:embed` impact how these template directories are laid out, and the workarounds are file rename in basis. The
impacts are noted below for information and future follow-up:

#### `go:embed` doesn't traverse hidden directories, but Rust projects include a hidden directory 
Our Rust examples use [Cargo][https://doc.rust-lang.org/cargo/reference/config.html] as a build tool. This stores
configuration in a hidden directory `.cargo`. As of Go 1.16, hidden directories are not yet supported with `go:embed`.
See https://github.com/golang/go/issues/43854

#### `go:embed` stops traversing at module boundaries, and TinyGo examples look like sub-modules
Go modules need to be built from the zip-file uploaded to the mirror. This implies `go:embed` must refer to the
current module. https://github.com/golang/go/issues/45197 explains embedding stops traversing when it encounters a
`go.mod` file, and there is no plan to change this.

This presents a challenge with TinyGo templates, which except for parameterization like `module {{ .Extension.Name }}`,
appear as normal go modules (ex `go.mod`) files. When `go:embed` encounters a `go.mod` file, it stops traversing. If we
didn't work around this, no TinyGo project would end up in the embedded filesystem.

The workaround is to rename the template `go.mod` to `go.mod_`, but this introduces a couple glitches:

* TinyGo extension templates imports appear in the root project after running `go mod tidy`
  * As of the writing, there is only one github.com/tetratelabs/proxy-wasm-go-sdk
* TinyGo extension templates can't be run from their directory without temporarily renaming `go.mod_` back to `go.mod`
  * To do this anyway, you'd need to fix any template variables like `module {{ .Extension.Name }}`

### Former solution
In the past, we embedded via [statik](https://github.com/rakyll/statik). This is a code generation solution which
encodes files into internal variables. Entry-points use an `http.FileSystem` to access the embedded files.

This solution worked well, except that it introduced build complexity. Code generation required an `init` phase in the
`Makefile`, as well lint exclusions. This implied steps for developers to remember and CI to invoke.

It also introduced maintenance and risk as the statik library stalled. This project ended up pinned to a personal fork
in order to work around unmerged pull requests.
```
replace github.com/rakyll/statik => github.com/yskopets/statik v0.1.8-0.20200501213002-c2d8dcc79889
```

Go 1.16's `go:embed` has limitations of its own, but brings with it shared understanding. While we may need workarounds
to some issues, those issues are understood by the go community. `go:embed` does not require our maintenance, nor has
any key person risk.
