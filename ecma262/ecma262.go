// Package ecma262 reproduces the single-character classes of the ECMAScript
// (JavaScript) regular expression language as rune predicates. Its reason
// for existing: OpenAPI pattern fields are ECMA-262, and the semantics do
// not carry over to Go's regexp verbatim — ECMA \w is ASCII-only, and ECMA
// \s is wider than Go's \s.
package ecma262

// IsWord reports whether r belongs to ECMAScript \w: ASCII letters, digits,
// or underscore. Matches Go's \w.
func IsWord(r rune) bool {
	return (r >= 'A' && r <= 'Z') ||
		(r >= 'a' && r <= 'z') ||
		(r >= '0' && r <= '9') ||
		r == '_'
}

// IsSpace reports whether r belongs to ECMAScript \s: ASCII \t, \n, \v, \f,
// \r and space; plus NBSP (U+00A0), U+1680, the en-quad through hair-space
// run (U+2000 to U+200A), line and paragraph separators (U+2028, U+2029),
// narrow no-break space (U+202F), medium mathematical space (U+205F),
// ideographic space (U+3000), and the byte order mark (U+FEFF) — all of
// which Go's \s rejects.
func IsSpace(r rune) bool {
	switch r {
	case '\t', '\n', '\v', '\f', '\r', ' ',
		0x00A0, 0x1680, 0x2028, 0x2029, 0x202F, 0x205F, 0x3000, 0xFEFF:
		return true
	}
	return r >= 0x2000 && r <= 0x200A
}

// Only reports whether s is non-empty and every rune satisfies allowed,
// matching the ^(...)+$ anchor shape of an OpenAPI pattern.
func Only(s string, allowed func(rune) bool) bool {
	if s == "" {
		return false
	}
	for _, r := range s {
		if !allowed(r) {
			return false
		}
	}
	return true
}
