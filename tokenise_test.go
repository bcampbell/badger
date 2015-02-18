package badger

import (
	"testing"
)

func cmpstrs(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i, _ := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

func TestTokenise(t *testing.T) {
	testData := []struct {
		in     string
		expect []string
	}{
		{"Hello There", []string{"hello", "there"}},
		{"Fred's hat", []string{"freds", "hat"}},
		{"for $10000000.", []string{"for", "10000000"}},
		{"a book-end!", []string{"a", "bookend"}},
		{"will it end in chaos?", []string{"will", "it", "end", "in", "chaos"}},
	}

	for _, foo := range testData {
		got := Tokenise(foo.in)
		if !cmpstrs(got, foo.expect) {
			t.Errorf("Tokenise(%q): expected %q, got %q", foo.in, foo.expect, got)
		}
	}
}
