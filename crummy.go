package crummy

import (
	"encoding/json"
	"fmt"
	"io"
	"reflect"
	"strings"
)

// Caveats:
// - have to store ptrs to structs
// - can only query on string and []string fields (but can store anything)
//
type Store struct {
	docs    map[string]interface{}
	docType reflect.Type
}

type DocSet map[string]struct{}

func NewStore(exampleDoc interface{}) *Store {
	store := &Store{
		docs:    make(map[string]interface{}),
		docType: reflect.TypeOf(exampleDoc),
	}

	fmt.Printf("New store docType=%s\n", store.docType)
	return store
}
func getString(doc interface{}, field string) string {
	v := reflect.ValueOf(doc)
	s := v.Elem()

	f := s.FieldByName(field)
	if !f.IsValid() {
		panic(field + " not found")
	}
	if f.Kind() != reflect.String {
		panic(field + " not a string")
	}

	return f.String()
}

func (store *Store) Count() int {
	return len(store.docs)
}

func (store *Store) Put(id string, doc interface{}) {
	store.docs[id] = doc
}

func (store *Store) FindAll() DocSet {
	matching := DocSet{}
	for id, _ := range store.docs {
		matching[id] = struct{}{}
	}
	return matching
}

func (store *Store) find(field string, cmp func(string) bool) DocSet {

	// resolve the field
	sf, ok := store.docType.FieldByName(field)
	if !ok {
		panic("couldn't resolve " + field)
	}

	matching := DocSet{}

	// string or []string?
	switch sf.Type.Kind() {
	case reflect.String:
		//
		for id, doc := range store.docs {
			s := reflect.ValueOf(doc).Elem() // get struct
			f := s.FieldByIndex(sf.Index)
			if cmp(f.String()) {
				matching[id] = struct{}{}
			}
		}
	case reflect.Slice:
		// it's []string
		for id, doc := range store.docs {
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

func (store *Store) FindExact(field, val string) DocSet {

	return store.find(field, func(foo string) bool {
		return foo == val
	})
}

func (store *Store) FindContaining(field, val string) DocSet {
	return store.find(field, func(foo string) bool {
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

func Load(in io.Reader, exampleDoc interface{}) (*Store, error) {

	store := NewStore(exampleDoc)

	// create a map of our desired doctype to unmarshal into
	inMapType := reflect.MapOf(reflect.TypeOf(""), reflect.PtrTo(store.docType))

	inMap := reflect.New(inMapType)

	dec := json.NewDecoder(in)
	err := dec.Decode(inMap.Interface())
	if err != nil {
		return nil, err
	}

	// now load the docs (of correct doctype) into the store
	for _, key := range reflect.Indirect(inMap).MapKeys() {
		val := reflect.Indirect(inMap).MapIndex(key)
		store.Put(key.String(), val.Interface())
	}

	return store, nil
}

func (store *Store) Write(out io.Writer) error {
	enc := json.NewEncoder(out)
	return enc.Encode(store.docs)
}
