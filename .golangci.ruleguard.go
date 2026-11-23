// Copyright func-e contributors
// SPDX-License-Identifier: Apache-2.0

// Package gorules is loaded by golangci-lint's ruleguard to fail `_ =`
// boilerplate that the std-error-handling preset already covers.
package gorules

import "github.com/quasilyte/go-ruleguard/dsl"

// errcheckBoilerplate fails `_ =` on calls the std-error-handling preset
// already ignores.
func errcheckBoilerplate(m dsl.Matcher) {
	m.Match(`_ = $x.Close()`).
		Report(`call $x.Close() directly; std-error-handling ignores this error`)
	m.Match(`_ = $x.Flush()`).
		Report(`call $x.Flush() directly; std-error-handling ignores this error`)
	m.Match(`_, _ = fmt.$f($*_)`).
		Where(m["f"].Text.Matches(`^F?[Pp]rint`)).
		Report(`call fmt.$f directly; std-error-handling ignores this error`)
	m.Match(`_ = os.$f($*_)`).
		Where(m["f"].Text.Matches(`^(Remove|RemoveAll|Setenv|Unsetenv)$`)).
		Report(`call os.$f directly; std-error-handling ignores this error`)
	m.Match(`defer func() { $x.Close() }()`).
		Report(`use defer $x.Close()`)
}

// testErrcheckBoilerplate fails `_, _ =` on calls the test errcheck exclusion
// already ignores.
func testErrcheckBoilerplate(m dsl.Matcher) {
	m.Match(`_, _ = $w.Write($*_)`).
		Where(m["w"].Type.Is("http.ResponseWriter") && m.File().Name.Matches(`_test\.go$`)).
		Report(`call $w.Write directly; the test errcheck rule ignores this error`)
}

// closureBoilerplate fails unnecessary `func()` wrappers around single calls.
func closureBoilerplate(m dsl.Matcher) {
	m.Match(`defer func() { $f() }()`).Where(m["f"].Type.Underlying().Is("func()")).
		Report(`use defer $f()`)
	m.Match(`t.Cleanup(func() { $f() })`).Where(m["f"].Type.Underlying().Is("func()")).
		Report(`use t.Cleanup($f)`)
}
