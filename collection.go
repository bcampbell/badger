package badger

import (
	"fmt"
	//	"reflect"
	"strings"
	"sync"
)

type DocID uint
type TermPos int

type termMap map[DocID][]TermPos

// map terms->docs->positions
type index map[string]termMap

type TermSplitter func(string) []string
type TermCleaner func(string) string

// return true if it's a stopword
type StopwordChecker func(string) bool

type FieldPolicy struct {
	SplitTerms TermSplitter
	CleanTerm  TermCleaner
	IsStopword StopwordChecker

	BruteForceMatch bool
}

func (policy *FieldPolicy) Cook(txt string) ([]string, []int) {
	pos := 0
	rawTerms := policy.SplitTerms(txt)
	terms := []string{}
	positions := []int{}
	for _, term := range rawTerms {
		term = policy.CleanTerm(term)
		if !policy.IsStopword(term) {
			terms = append(terms, term)
			positions = append(positions, pos)
		}
		pos++
	}
	return terms, positions
}

func defaultStopwordChecker(string) {
}

var defaultPolicy = FieldPolicy{
	SplitTerms:      strings.Fields,
	CleanTerm:       strings.ToLower,
	IsStopword:      func(string) bool { return false },
	BruteForceMatch: false,
}

type Collection struct {
	sync.RWMutex

	fields map[string]FieldPolicy
	// map fields to indices
	indices      map[string]index
	allDocs      DocSet
	DefaultField string // field to search by default (mainly for the benefit of the query parser)
}

func NewCollection() *Collection {
	coll := Collection{
		indices: map[string]index{},
		allDocs: DocSet{},
	}
	return &coll
}

func (coll *Collection) Count() int {
	coll.RLock()
	defer coll.RUnlock()
	return len(coll.allDocs)
}

func (coll *Collection) Remove(docID DocID) {
	panic("not implemented")

	coll.Lock()
	defer coll.Unlock()

	delete(coll.allDocs, docID)
}

/*
func (coll *Collection) Get(id string) interface{} {
	coll.RLock()
	defer coll.RUnlock()
	return coll.docs[id]
}
*/

func (coll *Collection) fieldPolicy(field string) FieldPolicy {
	policy, got := coll.fields[field]
	if !got {
		return defaultPolicy
	}
	return policy
}

func (coll *Collection) setFieldPolicy(field string, policy FieldPolicy) {
	coll.fields[field] = policy
}

func (coll *Collection) findAll() DocSet {
	all := DocSet{}
	for docID, _ := range coll.allDocs {
		all[docID] = struct{}{}
	}
	return all
}

// Find executes a query and fills out a slice containing the results.
func (coll *Collection) Find(q Query) DocSet {
	return nil
}

func (coll *Collection) IndexText(docID DocID, field string, txt string, startPos TermPos) TermPos {
	fmt.Printf("lock?\n")
	coll.Lock()
	defer coll.Unlock()

	fmt.Printf("locked\n")
	policy := coll.fieldPolicy(field)
	if _, got := coll.indices[field]; !got {
		coll.indices[field] = make(index)
	}
	idx := coll.indices[field]

	terms := policy.SplitTerms(txt)
	fmt.Printf("terms: %q\n", terms)
	for i, _ := range terms {
		terms[i] = policy.CleanTerm(terms[i])
	}

	pos := startPos
	for _, term := range terms {
		if !policy.IsStopword(term) {
			if _, got := idx[term]; !got {
				idx[term] = termMap{} // first encounter of this term
			}
			idx[term][docID] = append(idx[term][docID], pos)
		}
		pos++
	}

	coll.allDocs[docID] = struct{}{}
	return pos
}
