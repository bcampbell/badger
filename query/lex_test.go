package query

import (
	//	"fmt"
	"testing"
)

func same(a, b []token) bool {
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

func TestLexer(t *testing.T) {
	testData := []struct {
		input    string
		expected []token
	}{
		{
			"", []token{
				token{tokEOF, ""},
			},
		},
		{
			`tag: citrus`, []token{
				token{tokLit, "tag"},
				token{tokColon, ":"},
				token{tokLit, "citrus"},
				token{tokEOF, ""},
			},
		},
		{
			`tag:(lemon mango)`, []token{
				token{tokLit, "tag"},
				token{tokColon, ":"},
				token{tokLParen, "("},
				token{tokLit, "lemon"},
				token{tokLit, "mango"},
				token{tokRParen, ")"},
				token{tokEOF, ""},
			},
		},
		{
			`-tag:("lemon mango")`, []token{
				token{tokMinus, "-"},
				token{tokLit, "tag"},
				token{tokColon, ":"},
				token{tokLParen, "("},
				token{tokQuoted, `"lemon mango"`},
				token{tokRParen, ")"},
				token{tokEOF, ""},
			},
		},
		{
			`tag:(citrus -banana)`, []token{
				token{tokLit, "tag"},
				token{tokColon, ":"},

				token{tokLParen, "("},
				token{tokLit, "citrus"},
				token{tokMinus, "-"},
				token{tokLit, "banana"},
				token{tokRParen, ")"},
				token{tokEOF, ""},
			},
		},
		{
			`date:[2014-01-01 TO 2014-01-02]`, []token{
				token{tokLit, "date"},
				token{tokColon, ":"},
				token{tokLSq, "["},
				token{tokLit, "2014-01-01"},
				token{tokTo, "TO"},
				token{tokLit, "2014-01-02"},
				token{tokRSq, "]"},
				token{tokEOF, ""},
			},
		},
	}

	for _, data := range testData {
		got := []token{}
		lex := lex(data.input)
		for tok := range lex.tokens {
			got = append(got, tok)
		}

		//		fmt.Printf("Got %v, expected %v\n", got, data.expected)

		if !same(got, data.expected) {
			t.Errorf(`Lex '%s' failed: got %v expected %v`, data.input, got, data.expected)
		}
	}
}
