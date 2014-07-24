package crummy

import (
	"strings"
)

type queryKind uint

const (
	All queryKind = iota
	Exact
	Contains
	Range
	OR
	AND
	Diff
)

type Query struct {
	op          queryKind
	field       string
	val, val2   string
	left, right *Query
}

func NewAllQuery() *Query {
	return &Query{op: All}
}

func NewExactQuery(field, value string) *Query {
	return &Query{op: Exact, field: field, val: strings.ToLower(value)}
}

func NewContainsQuery(field, value string) *Query {
	return &Query{op: Contains, field: field, val: strings.ToLower(value)}
}

func NewORQuery(left, right *Query) *Query {
	return &Query{op: OR, left: left, right: right}
}

func NewANDQuery(left, right *Query) *Query {
	return &Query{op: AND, left: left, right: right}
}

func NewRangeQuery(field, start, end string) *Query {
	return &Query{op: Range, field: field, val: start, val2: end}
}

func (q *Query) String() string {
	switch q.op {
	case All:
		return "<all>"
	case Exact:
		return "<exact " + q.field + ":" + q.val + ">"
	case Contains:
		return "<contains " + q.field + ":" + q.val + ">"
	case OR:
		return "<OR>"
	case AND:
		return "<AND>"
	case Diff:
		return "<diff>"
	case Range:
		return "<range [" + q.val + " TO " + q.val2 + "]>"
	}
	return ""
}

func DumpTree(q *Query, indent int) string {
	out := ""
	if q == nil {
		return "<empty>\n"
	}
	for i := 0; i < indent; i++ {
		out += "  "
	}
	out += q.String() + "\n"
	if q.left != nil {
		out += DumpTree(q.left, indent+1)
	}
	if q.right != nil {
		out += DumpTree(q.right, indent+1)
	}
	return out
}

func (q *Query) perform(coll *Collection) docSet {
	switch q.op {
	case All:
		return coll.findAll()
	case Exact:
		return coll.findExact(q.field, q.val)
	case Contains:
		return coll.findContains(q.field, q.val)
	case Range:
		return coll.findRange(q.field, q.val, q.val2)
	case AND:
		a := q.left.perform(coll)
		b := q.right.perform(coll)
		return Intersect(a, b)
	case OR:
		a := q.left.perform(coll)
		b := q.right.perform(coll)
		return Union(a, b)
	case Diff:
		panic("not implemented yet")
	default:
		panic("not implemented yet")
	}
}
