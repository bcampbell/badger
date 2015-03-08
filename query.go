package badger

import (
	"fmt"
	"regexp"
	"strings"
)

type Query interface {
	perform(coll *Collection) DocSet
	String() string
}

type nilQuery struct {
}

// NewNilQuery returns a query which matches nothing
func NewNilQuery() Query {
	return &nilQuery{}
}

func (q *nilQuery) String() string {
	return "<NONE>"
}

func (q *nilQuery) perform(coll *Collection) DocSet {
	return DocSet{}
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

func (q *allQuery) perform(coll *Collection) DocSet {
	return coll.findAll()
}

//
type exactQuery struct {
	field  string
	values []string
}

// NewExactQuery returns a query to match a field exactly (but still case insensitively)
// If multiple values are given, then a match against _any_ of the values is sufficient for the
// document to be matched (ie the values can be considered to be separated by ORs rather than ANDs)
func NewExactQuery(field string, values ...string) Query {
	for i, _ := range values {
		values[i] = strings.ToLower(values[i])
	}
	return &exactQuery{field: field, values: values}
}

func (q *exactQuery) String() string {
	// TODO
	return q.field + ":blahblah"
}

func (q *exactQuery) perform(coll *Collection) DocSet {
	/*
		return coll.find(q.field, func(foo string) bool {
			foo = strings.ToLower(foo)
			for _, v := range q.values {
				if foo == v {
					return true
				}
			}
			return false
		})
	*/
	return DocSet{}
}

type containsQuery struct {
	field  string
	phrase string
}

// NewContainsQuery finds docs with field containing the phrase
func NewContainsQuery(field, phrase string) Query {
	return &containsQuery{field: field, phrase: phrase}
}

func (q *containsQuery) String() string {
	return fmt.Sprintf(`%s:%q`, q.field, q.phrase)
}

func (q *containsQuery) perform(coll *Collection) DocSet {
	policy := coll.fieldPolicy(q.field)
	idx, got := coll.indices[q.field]
	if !got {
		return DocSet{} // non-existant field?
	}

	// split phrase into terms
	terms, _ := policy.Cook(q.phrase)

	if len(terms) < 1 {
		return DocSet{} // nothing to search for
	}

	// HACK FOR NOW - just a single term
	matches := DocSet{}
	for docID, _ := range idx[terms[0]] {
		matches[docID] = struct{}{}
	}

	return matches
}

type notQuery struct {
	subQuery Query
}

// NewNOTQuery returns everything that doesn't match subquery q
func NewNOTQuery(q Query) Query {
	return &notQuery{subQuery: q}
}
func (q *notQuery) String() string {
	return "-" + q.subQuery.String()
}

func (q *notQuery) perform(coll *Collection) DocSet {
	out := coll.findAll()
	out.Subtract(q.subQuery.perform(coll))
	return out
}

type orQuery struct {
	left, right Query
}

// NewORQuery returns a boolean OR of two subqueries
func NewORQuery(left, right Query) Query {
	return &orQuery{left: left, right: right}
}
func (q *orQuery) String() string {
	return "(" + q.left.String() + " OR " + q.right.String() + ")"
}

func (q *orQuery) perform(coll *Collection) DocSet {
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
	return "(" + q.left.String() + " AND " + q.right.String() + ")"
}

func (q *andQuery) perform(coll *Collection) DocSet {
	a := q.left.perform(coll)
	b := q.right.perform(coll)
	return Intersect(a, b)
}

type rangeKind int

const (
	str rangeKind = iota
	date
)

type rangeQuery struct {
	kind               rangeKind
	field, first, last string
}

// NewRangeQuery returns a query to match docs with field values within
// inclusive range [first,last]
func NewRangeQuery(field, first, last string) Query {
	datePat := regexp.MustCompile(`^\d{4}-\d{2}-\d{2}$`)

	var kind rangeKind
	if first == "" && last == "" {
		return NewNilQuery()
	} else if first == "" && datePat.MatchString(last) {
		kind = date
	} else if last == "" && datePat.MatchString(first) {
		kind = date
	} else if datePat.MatchString(first) && datePat.MatchString(last) {
		kind = date
	}

	return &rangeQuery{kind: kind, field: field, first: strings.ToLower(first), last: strings.ToLower(last)}
}

func (q *rangeQuery) String() string {
	return q.field + ": [" + q.first + " TO " + q.last + "]"
}

var dateExtractPat *regexp.Regexp = regexp.MustCompile(`\d{4}-\d{2}-\d{2}`)

func (q *rangeQuery) perform(coll *Collection) DocSet {
	return DocSet{}
	/*
		if q.kind == str {
			// straight string compare
			// TODO: less-than/greater-than special cases
			return coll.find(q.field, func(foo string) bool {
				foo = strings.ToLower(foo)
				return foo >= q.first && foo <= q.last
			})
		} else {
			// date compare
			if q.first == "" {
				// less-than-or-equal-to
				return coll.find(q.field, func(foo string) bool {
					foo = dateExtractPat.FindString(foo)
					if foo == "" {
						return false
					}
					return foo <= q.last
				})
			}

			if q.last == "" {
				// greater-than-or-equal-to
				return coll.find(q.field, func(foo string) bool {
					foo = dateExtractPat.FindString(foo)
					if foo == "" {
						return false
					}
					return foo >= q.first
				})
			}
			// inclusive range compare
			return coll.find(q.field, func(foo string) bool {
				foo = dateExtractPat.FindString(foo)
				if foo == "" {
					return false
				}
				return foo >= q.first && foo <= q.last
			})
		}
	*/
}
