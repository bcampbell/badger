package query

import (
	"fmt"
	//"labix.org/v2/mgo"
	//"regexp"
	"semprini/crummy"
	"time"
)

type parser struct {
	tokens []token
	pos    int
}

/*
BNF syntax for query strings:

expr ::= andOp | orOp | group | range | [boolmod] [field ":"] lit
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

func Parse(q string, defaultField string) (*crummy.Query, error) {
	lex := lex(q)
	var tokens []token
	for tok := range lex.tokens {
		tokens = append(tokens, tok)
	}
	p := parser{tokens: tokens}
	return p.parseExpr(defaultField)
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
func (p *parser) parseExpr(defaultField string) (*crummy.Query, error) {
	if p.peek().typ == tokEOF {
		return nil, nil
	}

	// optional boolean modifier and field
	//boolMod := p.parseBoolMod()
	field := p.parseField()
	if field == "" {
		field = defaultField
	}

	var q *crummy.Query
	var err error
	tok := p.next()
	switch tok.typ {

	case tokLit:
		q = crummy.NewContainsQuery(field, tok.val)
	case tokQuoted:
		txt := string(tok.val[1 : len(tok.val)-1])
		q = crummy.NewContainsQuery(field, txt)
	case tokLSq:
		p.backup()
		start, end, err := p.parseRange()
		if err != nil {
			return nil, err
		}
		q = crummy.NewRangeQuery(field, start, end)
	case tokLParen:
		q, err = p.parseExpr(field)
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

	tok = p.next()

	if tok.typ == tokRParen {
		p.backup()
		return q, nil
	}
	if tok.typ == tokEOF {
		return q, nil
	}

	if tok.typ == tokOr {
		qr, err := p.parseExpr(field)
		if err != nil {
			return nil, err
		}
		return crummy.NewORQuery(q, qr), nil
	}

	if tok.typ != tokAnd {
		p.backup() // AND is optional :-)
	}
	qr, err := p.parseExpr(field)
	if err != nil {
		return nil, err
	}
	return crummy.NewANDQuery(q, qr), nil
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
func (p *parser) parseField() string {
	tok := p.next()
	if tok.typ != tokLit {
		// not a field
		p.backup()
		return ""
	}
	fieldName := tok.val

	tok = p.next()
	if tok.typ != tokColon {
		// oop, not a field after all!
		p.backup()
		p.backup()
		return ""
	}

	return fieldName
}

// expects "YYYY-MM-DD" form
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
