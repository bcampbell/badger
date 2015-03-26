package badger

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

type Query interface {
	perform(coll *Collection) docSet
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

func (q *nilQuery) perform(coll *Collection) docSet {
	return docSet{}
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
	if len(q.values) == 1 {
		return fmt.Sprintf(`%s:%s`, q.field, q.values[0])
	} else {
		return fmt.Sprintf(`%s: IN %v`, q.field, q.values)
	}
}

func (q *containsQuery) perform(coll *Collection) docSet {

	if _, got := coll.wholeWordFields[strings.ToLower(q.field)]; !got {
		// no whole-word check needed - just plain string search
		return coll.find(q.field, func(foo string) bool {
			foo = strings.ToLower(foo)
			for _, v := range q.values {
				if strings.Contains(foo, v) {
					return true
				}
			}
			return false
		})

	} else {
		// require whole-word matching (ie "tory" does not match "history")

		return coll.find(q.field, func(foo string) bool {
			/*
				// 1st pass - just do string search
				found := false
				foo = strings.ToLower(foo)
				for _, v := range q.values {
					if strings.Contains(foo, v) {
						found = true
					}
				}
				if !found {
					return false
				}
			*/

			// now do more rigorous check:
			searchSpace := Tokenise(foo)
			for _, v := range q.values {
				// the search phrase might tokenise into multiple terms
				searchTerms := Tokenise(v)
				for pos := 0; pos <= len(searchSpace)-len(searchTerms); pos++ {
					t := 0
					for ; t < len(searchTerms); t++ {
						if searchSpace[pos+t] != searchTerms[t] {
							break
						}
					}
					if t == len(searchTerms) {
						// got a full match!
						return true
					}
				}
			}
			return false
		})
	}

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

func (q *notQuery) perform(coll *Collection) docSet {
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
	return "(" + q.left.String() + " AND " + q.right.String() + ")"
}

func (q *andQuery) perform(coll *Collection) docSet {
	a := q.left.perform(coll)
	b := q.right.perform(coll)
	return Intersect(a, b)
}

const maxUint = ^uint(0)
const minUint = 0
const maxInt = int(maxUint >> 1)
const minInt = -maxInt - 1

var dateExtractPat *regexp.Regexp = regexp.MustCompile(`\d{4}-\d{2}-\d{2}`)

// NewRangeQuery returns a query to match docs with field values within
// inclusive range [first,last]
func NewRangeQuery(field, first, last string) Query {
	datePat := regexp.MustCompile(`^\d{4}-\d{2}-\d{2}$`)

	if first == "" && last == "" {
		return NewNilQuery()
	}
	if first == "" && datePat.MatchString(last) {
		return &dateRangeQuery{field, first, last}
	}
	if last == "" && datePat.MatchString(first) {
		return &dateRangeQuery{field, first, last}
	}
	if datePat.MatchString(first) && datePat.MatchString(last) {
		return &dateRangeQuery{field, first, last}
	}

	a, aErr := strconv.Atoi(first)
	b, bErr := strconv.Atoi(last)
	if first == "" && bErr == nil {
		return &intRangeQuery{field, minInt, b}
	}

	if aErr == nil && last == "" {
		return &intRangeQuery{field, a, maxInt}
	}

	if aErr == nil && bErr == nil {
		return &intRangeQuery{field, a, b}
	}

	return &strRangeQuery{field, strings.ToLower(first), strings.ToLower(last)}

}

// string range
type strRangeQuery struct {
	field, first, last string
}

func (q *strRangeQuery) String() string {
	return q.field + ": [" + q.first + " TO " + q.last + "]"
}
func (q *strRangeQuery) perform(coll *Collection) docSet {
	// straight string compare
	// TODO: less-than/greater-than special cases
	return coll.find(q.field, func(foo string) bool {
		foo = strings.ToLower(foo)
		return foo >= q.first && foo <= q.last
	})
}

// date range
type dateRangeQuery struct {
	field, first, last string
}

func (q *dateRangeQuery) String() string {
	return q.field + ": [" + q.first + " TO " + q.last + "]"
}

func (q *dateRangeQuery) perform(coll *Collection) docSet {
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

// inclusive integer range
type intRangeQuery struct {
	field       string
	first, last int
}

func (q *intRangeQuery) String() string {
	return fmt.Sprintf("%s: [%d TO %d]", q.field, q.first, q.last)
}

func (q *intRangeQuery) perform(coll *Collection) docSet {
	return coll.find(q.field, func(foo string) bool {
		v, err := strconv.Atoi(foo)
		if err != nil {
			return false
		}
		return v >= q.first && v <= q.last
	})
}
