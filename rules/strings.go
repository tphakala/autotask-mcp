//go:build ruleguard

package gorules

import "github.com/quasilyte/go-ruleguard/dsl"

// StringsLinesIteration detects manual line splitting patterns and suggests strings.Lines.
//
// Old pattern:
//
//	for _, line := range strings.Split(s, "\n") {
//	    process(line)
//	}
//
// New pattern (Go 1.24+):
//
//	for line := range strings.Lines(s) {
//	    process(line)
//	}
//
// Benefits:
//   - No intermediate slice allocation
//   - Handles both \n and \r\n line endings
//   - More memory efficient for large strings
//
// See: https://pkg.go.dev/strings#Lines
// See: https://pkg.go.dev/bytes#Lines
func StringsLinesIteration(m dsl.Matcher) {
	// Pattern: for _, line := range strings.Split(s, "\n")
	m.Match(
		`for $_, $line := range strings.Split($s, "\n") { $*body }`,
	).
		Report(`use for $line := range strings.Lines($s) instead of ranging over strings.Split($s, "\n") (Go 1.24+); note: Lines() handles both \n and \r\n`)

	// Pattern: for _, line := range strings.Split(s, "\r\n")
	m.Match(
		`for $_, $line := range strings.Split($s, "\r\n") { $*body }`,
	).
		Report(`use for $line := range strings.Lines($s) instead of ranging over strings.Split($s, "\r\n") (Go 1.24+)`)

	// Also detect bytes.Split for line iteration
	m.Match(
		`for $_, $line := range bytes.Split($s, []byte("\n")) { $*body }`,
	).
		Report(`use for $line := range bytes.Lines($s) instead of ranging over bytes.Split($s, []byte("\n")) (Go 1.24+)`)

	m.Match(
		`for $_, $line := range bytes.Split($s, []byte{'\n'}) { $*body }`,
	).
		Report(`use for $line := range bytes.Lines($s) instead of ranging over bytes.Split (Go 1.24+)`)
}

// StringsSplitIteration detects strings.Split used only for iteration
// and suggests strings.SplitSeq for better memory efficiency.
//
// Old pattern:
//
//	for _, part := range strings.Split(s, ",") {
//	    process(part)
//	}
//
// New pattern (Go 1.24+):
//
//	for part := range strings.SplitSeq(s, ",") {
//	    process(part)
//	}
//
// Benefits:
//   - No intermediate slice allocation
//   - Better for large strings with many parts
//   - Works with iterator composition
//
// Note: Only use SplitSeq when you're just iterating. If you need the slice
// result (e.g., to access by index or get length), keep using Split.
//
// See: https://pkg.go.dev/strings#SplitSeq
// See: https://pkg.go.dev/bytes#SplitSeq
func StringsSplitIteration(m dsl.Matcher) {
	// Pattern: for _, part := range strings.Split(s, sep)
	// Excluding newline separators which should use Lines() instead
	m.Match(
		`for $_, $part := range strings.Split($s, $sep) { $*body }`,
	).
		Where(!m["sep"].Text.Matches(`^"\\n"$`) && !m["sep"].Text.Matches(`^"\\r\\n"$`)).
		Report("use for $part := range strings.SplitSeq($s, $sep) to avoid intermediate slice allocation (Go 1.24+)")

	// bytes.Split pattern
	m.Match(
		`for $_, $part := range bytes.Split($s, $sep) { $*body }`,
	).
		Where(!m["sep"].Text.Matches(`\[\]byte\("\\n"\)`) && !m["sep"].Text.Matches(`\[\]byte\{.*\\n.*\}`)).
		Report("use for $part := range bytes.SplitSeq($s, $sep) to avoid intermediate slice allocation (Go 1.24+)")
}

// StringsFieldsIteration detects strings.Fields used only for iteration
// and suggests strings.FieldsSeq.
//
// Old pattern:
//
//	for _, field := range strings.Fields(s) {
//	    process(field)
//	}
//
// New pattern (Go 1.24+):
//
//	for field := range strings.FieldsSeq(s) {
//	    process(field)
//	}
//
// See: https://pkg.go.dev/strings#FieldsSeq
// See: https://pkg.go.dev/bytes#FieldsSeq
func StringsFieldsIteration(m dsl.Matcher) {
	m.Match(
		`for $_, $field := range strings.Fields($s) { $*body }`,
	).
		Report("use for $field := range strings.FieldsSeq($s) to avoid intermediate slice allocation (Go 1.24+)")

	m.Match(
		`for $_, $field := range bytes.Fields($s) { $*body }`,
	).
		Report("use for $field := range bytes.FieldsSeq($s) to avoid intermediate slice allocation (Go 1.24+)")
}

// StringsFieldsFuncIteration detects strings.FieldsFunc used only for iteration
// and suggests strings.FieldsFuncSeq.
//
// Old pattern:
//
//	for _, field := range strings.FieldsFunc(s, f) {
//	    process(field)
//	}
//
// New pattern (Go 1.24+):
//
//	for field := range strings.FieldsFuncSeq(s, f) {
//	    process(field)
//	}
//
// See: https://pkg.go.dev/strings#FieldsFuncSeq
// See: https://pkg.go.dev/bytes#FieldsFuncSeq
func StringsFieldsFuncIteration(m dsl.Matcher) {
	m.Match(
		`for $_, $field := range strings.FieldsFunc($s, $f) { $*body }`,
	).
		Report("use for $field := range strings.FieldsFuncSeq($s, $f) to avoid intermediate slice allocation (Go 1.24+)")

	m.Match(
		`for $_, $field := range bytes.FieldsFunc($s, $f) { $*body }`,
	).
		Report("use for $field := range bytes.FieldsFuncSeq($s, $f) to avoid intermediate slice allocation (Go 1.24+)")
}

// StringsCutLast detects strings.LastIndex / bytes.LastIndex followed by manual
// slicing or an index check, and suggests strings.CutLast / bytes.CutLast.
//
// Old patterns:
//
//	dir := path[:strings.LastIndex(path, "/")]
//	base := path[strings.LastIndex(path, "/")+1:]
//	if i := strings.LastIndex(s, sep); i >= 0 {
//	    before, after := s[:i], s[i+len(sep):]
//	}
//
// New pattern (Go 1.27+):
//
//	before, after, found := strings.CutLast(s, sep)
//
// CutLast returns (s, "", false) when sep is absent (bytes.CutLast returns
// (s, nil, false)). The slicing idioms behave differently in that case:
// s[:LastIndex] panics on -1, and s[LastIndex+1:] yields the whole string. Check
// the found result when the not-found case matters.
//
// The +1 slicing form fires only for a one-character literal separator, since
// with a longer separator it keeps part of it and CutLast would not. The
// index-check forms fire only when the guarded code slices s at the index
// (s[:i], s[i:], s[i+n:]); an index used for anything else is not a CutLast
// candidate.
//
// See: https://pkg.go.dev/strings#CutLast
// See: https://pkg.go.dev/bytes#CutLast
func StringsCutLast(m dsl.Matcher) {
	// Slicing around the last separator.
	m.Match(
		`$s[:strings.LastIndex($s, $sep)]`,
		`$s[strings.LastIndex($s, $sep)+len($sep):]`,
	).
		Report("use before, after, found := strings.CutLast($s, $sep) instead of slicing around strings.LastIndex (Go 1.27+); when $sep is absent the slicing idiom yields all of $s but CutLast yields after == \"\", so check found")

	m.Match(
		`$s[:bytes.LastIndex($s, $sep)]`,
		`$s[bytes.LastIndex($s, $sep)+len($sep):]`,
	).
		Report("use before, after, found := bytes.CutLast($s, $sep) instead of slicing around bytes.LastIndex (Go 1.27+); when $sep is absent the slicing idiom yields all of $s but CutLast yields after == nil, so check found")

	// The +1 form skips exactly one byte, so it is a CutLast candidate only when
	// the separator is a one-character literal; with "::" it would keep a ":".
	m.Match(
		`$s[strings.LastIndex($s, $sep)+1:]`,
	).
		Where(m["sep"].Text.Matches(`^"(\\.|[^"\\])"$`)).
		Report("use before, after, found := strings.CutLast($s, $sep) instead of slicing around strings.LastIndex (Go 1.27+); when $sep is absent the slicing idiom yields all of $s but CutLast yields after == \"\", so check found")

	m.Match(
		`$s[bytes.LastIndex($s, $sep)+1:]`,
	).
		Where(m["sep"].Text.Matches(`^\[\]byte\("(\\.|[^"\\])"\)$`) || m["sep"].Text.Matches(`^\[\]byte\{'(\\.|[^'\\])'\}$`)).
		Report("use before, after, found := bytes.CutLast($s, $sep) instead of slicing around bytes.LastIndex (Go 1.27+); when $sep is absent the slicing idiom yields all of $s but CutLast yields after == nil, so check found")

	// Index check followed by manual slicing in the body.
	m.Match(
		`if $i := strings.LastIndex($s, $sep); $i >= 0 { $*body }`,
		`if $i := strings.LastIndex($s, $sep); $i != -1 { $*body }`,
		`$i := strings.LastIndex($s, $sep); if $i < 0 { $*_ }; $*body`,
		`$i := strings.LastIndex($s, $sep); if $i == -1 { $*_ }; $*body`,
	).
		Where(m["body"].Contains(`$s[:$i]`) || m["body"].Contains(`$s[$i:]`) || m["body"].Contains(`$s[$i+$_:]`)).
		Report("use before, after, found := strings.CutLast($s, $sep) instead of checking strings.LastIndex and slicing by hand (Go 1.27+)")

	m.Match(
		`if $i := bytes.LastIndex($s, $sep); $i >= 0 { $*body }`,
		`if $i := bytes.LastIndex($s, $sep); $i != -1 { $*body }`,
		`$i := bytes.LastIndex($s, $sep); if $i < 0 { $*_ }; $*body`,
		`$i := bytes.LastIndex($s, $sep); if $i == -1 { $*_ }; $*body`,
	).
		Where(m["body"].Contains(`$s[:$i]`) || m["body"].Contains(`$s[$i:]`) || m["body"].Contains(`$s[$i+$_:]`)).
		Report("use before, after, found := bytes.CutLast($s, $sep) instead of checking bytes.LastIndex and slicing by hand (Go 1.27+)")
}
