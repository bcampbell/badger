package crummy

import (
	"encoding/json"
	"fmt"
	"io"
	"reflect"
	"strings"
)

var version string = "1"

// Caveats:
// - have to coll ptrs to structs
// - can only query on string and []string fields (but can store anything)
//
type Collection struct {
	docs    map[string]interface{}
	docType reflect.Type
}

type DocSet map[string]struct{}

func NewCollection(exampleDoc interface{}) *Collection {
	coll := &Collection{
		docs:    make(map[string]interface{}),
		docType: reflect.TypeOf(exampleDoc),
	}

	return coll
}

func (coll *Collection) Count() int {
	return len(coll.docs)
}

func (coll *Collection) Put(id string, doc interface{}) {
	coll.docs[id] = doc
}

func (coll *Collection) FindAll() DocSet {
	matching := DocSet{}
	for id, _ := range coll.docs {
		matching[id] = struct{}{}
	}
	return matching
}

func (coll *Collection) find(field string, cmp func(string) bool) DocSet {

	// resolve the field
	sf, ok := coll.docType.FieldByName(field)
	if !ok {
		panic("couldn't resolve " + field)
	}

	matching := DocSet{}

	// string or []string?
	switch sf.Type.Kind() {
	case reflect.String:
		//
		for id, doc := range coll.docs {
			s := reflect.ValueOf(doc).Elem() // get struct
			f := s.FieldByIndex(sf.Index)
			if cmp(f.String()) {
				matching[id] = struct{}{}
			}
		}
	case reflect.Slice:
		// it's []string
		for id, doc := range coll.docs {
			s := reflect.ValueOf(doc).Elem() // get struct
			f := s.FieldByIndex(sf.Index)    // get slice
			// check each item in the slice
			for idx := 0; idx < f.Len(); idx++ {
				if cmp(f.Index(idx).String()) {
					matching[id] = struct{}{}
					break
				}
			}
		}
	default:
		panic("can only query string and []string fields")
	}
	return matching
}

func (coll *Collection) FindRange(field, start, end string) DocSet {

	return coll.find(field, func(foo string) bool {
		return foo >= start && foo <= end
	})
}
func (coll *Collection) FindExact(field, val string) DocSet {

	return coll.find(field, func(foo string) bool {
		return foo == val
	})
}

func (coll *Collection) FindContains(field, val string) DocSet {
	return coll.find(field, func(foo string) bool {
		return strings.Contains(foo, val)
	})
}

func Union(a, b DocSet) DocSet {
	out := DocSet{}
	var id string
	for id, _ = range a {
		out[id] = struct{}{}
	}
	for id, _ = range b {
		out[id] = struct{}{}
	}
	return out
}

func Intersect(a, b DocSet) DocSet {
	out := DocSet{}
	var id string
	for id, _ = range a {
		if _, got := b[id]; got {
			out[id] = struct{}{}
		}
	}
	return out
}

type header struct {
	Version string
	DocType string
	Count   int
}

func Read(in io.Reader, exampleDoc interface{}) (*Collection, error) {

	coll := NewCollection(exampleDoc)

	dec := json.NewDecoder(in)
	var hdr header

	var err error
	err = dec.Decode(&hdr)
	if err != nil {
		return nil, err
	}

	if hdr.Version != version {
		return nil, fmt.Errorf("invalid version")
	}

	if hdr.DocType != coll.docType.String() {
		return nil, fmt.Errorf("doc type mismatch (expected '%s', got '%s')", coll.docType.String(), hdr.DocType)
	}

	inType := reflect.PtrTo(coll.docType)
	for i := 0; i < hdr.Count; i++ {
		var key string
		doc := reflect.New(inType)
		err = dec.Decode(&key)
		if err != nil {
			return nil, err
		}
		err = dec.Decode(doc.Interface())
		if err != nil {
			return nil, err
		}
		coll.Put(key, doc.Interface())
	}

	return coll, nil
}

func (coll *Collection) Write(out io.Writer) error {
	var err error
	enc := json.NewEncoder(out)

	hdr := header{Version: version,
		DocType: coll.docType.String(),
		Count:   len(coll.docs)}
	err = enc.Encode(hdr)
	if err != nil {
		return err
	}
	for key, doc := range coll.docs {
		err = enc.Encode(key)
		if err != nil {
			return err
		}
		err = enc.Encode(doc)
		if err != nil {
			return err
		}
	}
	return nil
}
