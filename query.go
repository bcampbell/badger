package badger

import (
	"strings"
)

type queryKind uint

const (
	All queryKind = iota
	Exact
	Contains
	Range
	ExactIn
	OR
	AND
	Diff
)

type Query interface {
	perform(coll *Collection) docSet
	String() string
}

type allQuery struct {
}

// NewAllQuery returns a query which matches all docs
func NewAllQuery() Query {
	return &allQuery{}
}

func (q *allQuery) String() string {
	return "<ALL>"
}

func (q *allQuery) perform(coll *Collection) docSet {
	return coll.findAll()
}

//
type exactQuery struct {
	field  string
	values []string
}

// NewExactQuery returns a query to match a field exactly
func NewExactQuery(field, value string) Query {
	return &exactQuery{field: field, values: []string{strings.ToLower(value)}}
}

func (q *exactQuery) String() string {
	// TODO
	return q.field + ":blahblah"
}

func (q *exactQuery) perform(coll *Collection) docSet {
	return coll.find(q.field, func(foo string) bool {
		foo = strings.ToLower(foo)
		for _, v := range q.values {
			if foo == v {
				return true
			}
		}
		return false
	})
}

type containsQuery struct {
	field  string
	values []string
}

// NewContainsQuery finds docs with field containing the value
func NewContainsQuery(field, value string) Query {
	return &containsQuery{field: field, values: []string{strings.ToLower(value)}}
}

func (q *containsQuery) String() string {
	// TODO
	return q.field + ":blahblah"
}
func (q *containsQuery) perform(coll *Collection) docSet {
	return coll.find(q.field, func(foo string) bool {
		for _, v := range q.values {
			foo = strings.ToLower(foo)
			if strings.Contains(foo, v) {
				return true
			}
		}

		return false
	})
}

type orQuery struct {
	left, right Query
}

// NewORQuery returns a boolean OR of two subqueries
func NewORQuery(left, right Query) Query {
	return &orQuery{left: left, right: right}
}
func (q *orQuery) String() string {
	return "(" + q.left.String() + "OR" + q.right.String() + ")"
}

func (q *orQuery) perform(coll *Collection) docSet {
	a := q.left.perform(coll)
	b := q.right.perform(coll)
	return Union(a, b)
}

type andQuery struct {
	left, right Query
}

// NewANDQuery returns a boolean AND of two subqueries
func NewANDQuery(left, right Query) Query {
	return &andQuery{left: left, right: right}
}

func (q *andQuery) String() string {
	return "(" + q.left.String() + "AND" + q.right.String() + ")"
}

func (q *andQuery) perform(coll *Collection) docSet {
	a := q.left.perform(coll)
	b := q.right.perform(coll)
	return Intersect(a, b)
}

type rangeQuery struct {
	field, first, last string
}

// NewRangeQuery returns a query to match docs with field values within
// inclusive range [first,last]
func NewRangeQuery(field, first, last string) Query {
	return &rangeQuery{field: field, first: strings.ToLower(first), last: strings.ToLower(last)}
}

func (q *rangeQuery) String() string {
	return q.field + ": [" + q.first + " TO " + q.last + "]"
}

func (q *rangeQuery) perform(coll *Collection) docSet {
	return coll.find(q.field, func(foo string) bool {
		foo = strings.ToLower(foo)
		// TODO: handle dates better!
		return foo >= q.first && foo <= q.last
	})
}
