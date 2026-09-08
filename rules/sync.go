//go:build ruleguard

// Package gorules defines custom linter rules for Go modernization.
package gorules

import "github.com/quasilyte/go-ruleguard/dsl"

// WaitGroupGo detects the old sync.WaitGroup pattern and suggests using Go 1.25's wg.Go().
//
// The old pattern:
//
//	wg.Add(1)
//	go func() {
//	    defer wg.Done()
//	    doSomething()
//	}()
//
// Can be simplified to:
//
//	wg.Go(func() {
//	    doSomething()
//	})
//
// Benefits:
//   - Cleaner, less error-prone (no Add/Done mismatch)
//   - Single function call
//   - Automatic panic handling
//
// See: https://pkg.go.dev/sync#WaitGroup.Go
func WaitGroupGo(m dsl.Matcher) {
	// Pattern 1: wg.Add(1) followed by go func() with defer wg.Done()
	// This matches when the defer is the first statement
	m.Match(
		`$wg.Add(1); go func() { defer $wg.Done(); $*body }()`,
	).
		Where(m["wg"].Type.Is("*sync.WaitGroup") || m["wg"].Type.Is("sync.WaitGroup")).
		Report("use $wg.Go(func() { $body }) instead of manual Add/Done pattern (Go 1.25+)").
		Suggest("$wg.Go(func() { $body })")

	// Pattern 2: Same but with pointer receiver explicitly
	m.Match(
		`$wg.Add(1); go func() { defer $wg.Done(); $*body }()`,
	).
		Where(m["wg"].Type.Underlying().Is("sync.WaitGroup")).
		Report("use $wg.Go(func() { $body }) instead of manual Add/Done pattern (Go 1.25+)").
		Suggest("$wg.Go(func() { $body })")

	// Pattern 3: When wg is passed by reference to the closure
	m.Match(
		`$wg.Add(1); go func($param $typ) { defer $param.Done(); $*body }($wg)`,
		`$wg.Add(1); go func($param $typ) { defer $param.Done(); $*body }(&$wg)`,
	).
		Report("use $wg.Go(func() { $body }) instead of manual Add/Done pattern (Go 1.25+)")
}

// AtomicTypes detects the primitive sync/atomic functions on plain integers and
// pointers and suggests the typed atomic wrappers (atomic.Int64, atomic.Pointer,
// and so on). This mirrors the atomictypes modernizer that go fix gained in Go 1.27.
//
// Old pattern:
//
//	var hits int64
//	atomic.AddInt64(&hits, 1)
//	n := atomic.LoadInt64(&hits)
//
// New pattern (types available since Go 1.19):
//
//	var hits atomic.Int64
//	hits.Add(1)
//	n := hits.Load()
//
// Benefits:
//   - The compiler rejects non-atomic reads and writes of the field
//   - 64-bit alignment is guaranteed on 32-bit platforms
//   - The intent is visible in the type, not only at each call site
//
// See: https://pkg.go.dev/sync/atomic#Int64
// See: https://pkg.go.dev/sync/atomic#Pointer
func AtomicTypes(m dsl.Matcher) {
	m.Match(
		`atomic.$fn(&$x, $*_)`,
	).
		Where(m["fn"].Text.Matches(`^(Add|Load|Store|Swap|CompareAndSwap|And|Or)(Int32|Int64|Uint32|Uint64|Uintptr)$`)).
		Report("declare $x as the matching atomic type (atomic.Int64, atomic.Uint32, ...) and call its methods instead of the atomic.$fn function on &$x; typed atomics forbid non-atomic access and fix 32-bit alignment; not applicable if $x is marshaled or crosses an ABI boundary")

	m.Match(
		`atomic.$fn(&$x, $*_)`,
	).
		Where(m["fn"].Text.Matches(`^(Load|Store|Swap|CompareAndSwap)Pointer$`)).
		Report("declare $x as atomic.Pointer[T] and call its methods instead of the atomic.$fn function on &$x; the generic wrapper removes the unsafe.Pointer casts")
}
