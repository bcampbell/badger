package query

import (
	"semprini/crummy"
)

type nodeKind uint

const (
	Exact nodeKind = iota
	Contains
	Range
	Union
	Intersect
	Diff
)

type Query struct {
	op        nodeKind
	field     string
	val, val2 string
	left      *Query
	right     *Query
}

func (q *Query) String() string {
	switch q.op {
	case Exact:
		return "<exact " + q.field + ":" + q.val + ">"
	case Contains:
		return "<contains " + q.field + ":" + q.val + ">"
	case Union:
		return "<union>"
	case Intersect:
		return "<intersect>"
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

func (q *Query) Perform(coll *crummy.Collection) crummy.DocSet {
	switch q.op {
	case Exact:
		return coll.FindExact(q.field, q.val)
	case Contains:
		return coll.FindContains(q.field, q.val)
	case Range:
		return coll.FindRange(q.field, q.val, q.val2)
	case Intersect:
		a := q.left.Perform(coll)
		b := q.right.Perform(coll)
		return crummy.Intersect(a, b)
	case Union:
		a := q.left.Perform(coll)
		b := q.right.Perform(coll)
		return crummy.Union(a, b)
	case Diff:
		panic("not implemented yet")
	default:
		panic("not implemented yet")
	}
}
