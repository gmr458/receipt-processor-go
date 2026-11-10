package receipt

import (
	"unicode"

	"github.com/gmr458/receipt-processor-go/ecma262"
)

func isAlphanumeric(r rune) bool {
	return unicode.IsLetter(r) || unicode.IsDigit(r)
}

func isOdd(n int) bool {
	return n%2 != 0
}

// isShortDescriptionRune reports whether r may appear in an item's
// shortDescription: an ECMA-262 [\w\s-] character (the api.yml pattern).
func isShortDescriptionRune(r rune) bool {
	return ecma262.IsWord(r) || ecma262.IsSpace(r) || r == '-'
}

// isRetailerRune reports whether r may appear in a retailer name: the
// shortDescription set plus '&'.
func isRetailerRune(r rune) bool {
	return isShortDescriptionRune(r) || r == '&'
}
