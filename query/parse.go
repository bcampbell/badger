package query

import (
	"fmt"
	//"labix.org/v2/mgo"
	//"regexp"
	"github.com/bcampbell/badger"
	"strings"
	//	"time"
)

type parser struct {
	tokens []token
	pos    int
}

/*
BNF syntax for query strings:
expr ::= andOp | orOp | group | range | ["="] lit | field ":" expr | [boolmod] expr
andOp ::= expr expr | expr "AND" expr
orOp ::= expr "OR" expr
notOp ::= "NOT" expr
group ::= "(" expr ")"
range ::= "[" [start] "TO" [end] "]"

lit ::= string | quotedstring | doublequotedstring

string ::= /\S+/
quotedstring ::= /'(.*?)'/
doublequotedstring ::= /"(.*?)"/

*/

func Parse(q string, validFields []string, defaultField string) (badger.Query, error) {
	lex := lex(q)
	var tokens []token
	for tok := range lex.tokens {
		tokens = append(tokens, tok)
	}
	p := parser{tokens: tokens}
	return p.parseExpr(defaultField, validFields)
}

func (p *parser) peek() token {
	if p.pos < len(p.tokens) {
		tok := p.tokens[p.pos]
		return tok
	}
	return token{typ: tokEOF}
}

func (p *parser) backup() {
	p.pos -= 1
}

func (p *parser) next() token {
	if p.pos < len(p.tokens) {
		tok := p.tokens[p.pos]
		p.pos += 1
		return tok
	}
	p.pos += 1 // to make sure backup() works
	return token{typ: tokEOF}
}

// starting point
// BNF:
//     query ::= expr | expr query
func (p *parser) parseExpr(defaultField string, validFields []string) (badger.Query, error) {
	if p.peek().typ == tokEOF {
		return nil, nil
	}

	// optional boolean modifier (default tokPlus)
	boolMod := p.parseBoolMod()

	// optional field
	field, err := p.parseField(validFields)
	if err != nil {
		return nil, err
	}

	if field == "" {
		field = defaultField
	}

	var q badger.Query
	tok := p.next()
	switch tok.typ {

	case tokEquals:
		// exact match
		tok = p.next()
		if tok.typ == tokLit {
			q = badger.NewExactQuery(field, tok.val)
		} else if tok.typ == tokQuoted {
			txt := string(tok.val[1 : len(tok.val)-1])
			q = badger.NewExactQuery(field, txt)
		} else {
			return nil, fmt.Errorf("expected term directly after '=', but got '%s'", tok.val)
		}

	case tokLit:
		q = badger.NewContainsQuery(field, tok.val)
	case tokQuoted:
		txt := string(tok.val[1 : len(tok.val)-1])
		q = badger.NewContainsQuery(field, txt)
	case tokLSq:
		p.backup()
		start, end, err := p.parseRange()
		if err != nil {
			return nil, err
		}
		q = badger.NewRangeQuery(field, start, end)
	case tokLParen:
		q, err = p.parseExpr(field, validFields)
		if err != nil {
			return nil, err
		}
		fluff := p.next()
		if fluff.typ != tokRParen {
			return nil, fmt.Errorf("expected ), got %s", fluff)
		}

	default:
		return nil, fmt.Errorf("unexpected: %s", tok)
	}

	//
	if boolMod == tokMinus {
		q = badger.NewNOTQuery(q)
	}

	tok = p.next()

	if tok.typ == tokRParen {
		p.backup()
		return q, nil
	}
	if tok.typ == tokEOF {
		return q, nil
	}

	if tok.typ == tokOr {
		qr, err := p.parseExpr(defaultField, validFields)
		if err != nil {
			return nil, err
		}
		return badger.NewORQuery(q, qr), nil
	}

	if tok.typ != tokAnd {
		p.backup() // AND is optional :-)
	}
	qr, err := p.parseExpr(defaultField, validFields)
	if err != nil {
		return nil, err
	}
	return badger.NewANDQuery(q, qr), nil
}

// parse (optional) boolean modifier
func (p *parser) parseBoolMod() tokType {
	tok := p.next()
	if tok.typ == tokMinus || tok.typ == tokPlus {
		return tok.typ
	}
	p.backup()
	return tokPlus
}

// parse (optional) field specifier
// [ lit ":" ]
// returns field name or nil if not a field
func (p *parser) parseField(validFields []string) (string, error) {
	tok := p.next()
	if tok.typ != tokLit {
		// not a field
		p.backup()
		return "", nil
	}
	field := tok.val

	tok = p.next()
	if tok.typ != tokColon {
		// oop, not a field after all!
		p.backup()
		p.backup()
		return "", nil
	}

	// check against valid fields
	field = strings.ToLower(field)
	for _, f := range validFields {
		if strings.ToLower(f) == field {
			return field, nil // it's OK
		}
	}
	return "", fmt.Errorf("unknown field '%s'", field)
}

// expects "YYYY-MM-DD" form
/*
func (p *parser) parseDate() (time.Time, error) {
	tok := p.next()
	if tok.typ != tokLit {
		return time.Time{}, fmt.Errorf("expected date, got %s", tok)
	}
	t, err := time.Parse("2006-01-02", tok.val)
	if err != nil {
		return time.Time{}, fmt.Errorf("expected YYYY-MM-DD, got '%s' (%s)", tok.val, err)
	}
	return t, nil
}
*/

// BNF:
//     range ::= "[" [start] "TO" [end] "]"
func (p *parser) parseRange() (string, string, error) {
	tok := p.next()
	if tok.typ != tokLSq {
		return "", "", fmt.Errorf("expected [, got %s", tok)
	}
	var start, end string

	tok = p.next()
	switch tok.typ {
	case tokLit:
		start = tok.val
	case tokQuoted:
		start = string(tok.val[1 : len(tok.val)-1])
	case tokTo:
		p.backup()
		// empty start
	default:
		return "", "", fmt.Errorf("unexpected: %s", tok)
	}

	tok = p.next()
	if tok.typ != tokTo {
		return "", "", fmt.Errorf("expected TO, got %s", tok)
	}

	tok = p.next()
	switch tok.typ {
	case tokLit:
		end = tok.val
	case tokQuoted:
		end = string(tok.val[1 : len(tok.val)-1])
	case tokRSq:
		p.backup() // empty end value
	default:
		return "", "", fmt.Errorf("unexpected: %s", tok)
	}

	tok = p.next()
	if tok.typ != tokRSq {
		return "", "", fmt.Errorf("expected ], got %s", tok)
	}

	if start == "" && end == "" {
		return "", "", fmt.Errorf("empty range")
	}

	return start, end, nil
}
