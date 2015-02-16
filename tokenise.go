package badger

import (
	"strings"
	"unicode"
	"unicode/utf8"
)

func stripPunc(s string) string {
	buf := make([]byte, len(s))
	pos := 0
	for i := 0; i < len(s); {
		r, size := utf8.DecodeRuneInString(s[i:])
		i += size
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			utf8.EncodeRune(buf[pos:], r)
			pos += size
		}
	}
	return string(buf[:pos])
}

// simple tokenising fn.
// eg:
// "Hello, there!" => "hello" "there"
// "Bob's hat!" => "bobs" "hat"
func Tokenise(txt string) []string {
	txt = strings.ToLower(txt)
	foo := strings.Fields(txt)
	for i, _ := range foo {
		foo[i] = stripPunc(foo[i])
	}
	return foo
}
