package query

import (
	"fmt"
	"strings"
	"unicode"
	"unicode/utf8"
)

type tokType int

const (
	tokEOF tokType = iota
	tokError
	tokLit
	tokQuoted
	tokPlus
	tokMinus
	tokColon
	tokOr
	tokAnd
	tokLParen
	tokRParen
	tokLSq
	tokRSq
	tokTo
)

// some single-rune tokens
var singles = map[rune]tokType{
	'(': tokLParen,
	')': tokRParen,
	'[': tokLSq,
	']': tokRSq,
	':': tokColon,
	'+': tokPlus,
	'-': tokMinus,
}

type token struct {
	typ tokType
	val string
}

func (tok token) String() string {

	tokTypes := map[tokType]string{
		tokEOF:    "eof",
		tokError:  "error",
		tokLit:    "lit",
		tokQuoted: "quoted",
		tokOr:     "or",
		tokAnd:    "and",
		tokTo:     "to",
		tokPlus:   "plus",
		tokMinus:  "minus",
		tokColon:  "colon",
		tokLParen: "lparen",
		tokRParen: "rparen",
		tokLSq:    "lsq",
		tokRSq:    "rsq",
	}
	return fmt.Sprintf("%s[%s]", tokTypes[tok.typ], tok.val)
}

type stateFn func(*lexer) stateFn

type lexer struct {
	input   string
	tokens  chan token
	pos     int
	prevpos int
	start   int
}

func lex(input string) *lexer {
	l := &lexer{
		input:  input,
		tokens: make(chan token),
	}
	go l.run()
	return l
}

func (l *lexer) run() {
	for state := lexDefault; state != nil; {
		state = state(l)
	}
	close(l.tokens)
}

func (l *lexer) next() rune {
	l.prevpos = l.pos
	if l.eof() {
		return 0
	}
	r, w := utf8.DecodeRuneInString(l.input[l.pos:])
	l.pos += w
	return r
}
func (l *lexer) eof() bool {
	return l.pos >= len(l.input)
}

func (l *lexer) ignore() {
	l.start = l.pos
}
func (l *lexer) backup() {
	l.pos = l.prevpos
}
func (l *lexer) peek() rune {
	r := l.next()
	l.backup()
	return r
}

func (l *lexer) emit(t tokType) {
	l.tokens <- token{t, l.input[l.start:l.pos]}
	l.start = l.pos
}

func lexDefault(l *lexer) stateFn {
	// skip space
	for {
		if l.eof() {
			l.emit(tokEOF)
			return nil
		}
		r := l.next()
		if !unicode.IsSpace(r) {
			l.backup()
			l.ignore()

			if typ, got := singles[r]; got {
				l.next()
				l.emit(typ)
				return lexDefault
			}

			if r == '"' || r == '\'' {
				return lexQuoted
			}

			return lexText
		}
	}
}

func lexText(l *lexer) stateFn {
	for {
		if l.eof() {
			break
		}
		r := l.next()
		if unicode.IsSpace(r) || strings.ContainsRune("():[]", r) {
			l.backup()
			break
		}
	}

	switch l.input[l.start:l.pos] {
	case "OR":
		l.emit(tokOr)
	case "AND":
		l.emit(tokAnd)
	case "TO":
		l.emit(tokTo)
	default:
		l.emit(tokLit)
	}

	return lexDefault
}

func lexQuoted(l *lexer) stateFn {
	q := l.next()
	for {
		if l.eof() {
			l.emit(tokError)
			return nil
		}
		r := l.next()
		if r == q {
			break
		}
	}
	l.emit(tokQuoted)
	return lexDefault
}
