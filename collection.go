package badger

import (
	"encoding/gob"
	"fmt"
	"io"
	"reflect"
	"strings"
)

var magicCookie = []byte{'b', 'a', 'd', '1'}

// Collection holds a set of documents, all of the same type.
// Caveats:
// - have to store ptrs to structs
// - can only query on string and []string fields (but can store anything)
//
type Collection struct {
	docs         map[string]interface{}
	docType      reflect.Type
	DefaultField string // field to search by default (mainly for the benefit of the query parser)
}

// NewCollection initialises a collection for holding documents of
// same type as referenceDoc.
// The contents of referenceDoc are unimportant - a zero object is
// fine. Only it's type is used.
func NewCollection(referenceDoc interface{}) *Collection {
	coll := &Collection{
		docs:    make(map[string]interface{}),
		docType: reflect.TypeOf(referenceDoc),
	}

	return coll
}

func (coll *Collection) Count() int {
	return len(coll.docs)
}

// ValidField returns true if field is a valid one (case insenstive)
func (coll *Collection) ValidFields() []string {
	fields := []string{}
	for i := 0; i < coll.docType.NumField(); i++ {
		fields = append(fields, coll.docType.Field(i).Name)
	}
	return fields
}

func (coll *Collection) Put(id string, doc interface{}) {
	coll.docs[id] = doc
}

func (coll *Collection) Get(id string) interface{} {
	return coll.docs[id]
}

func (coll *Collection) findAll() docSet {
	matching := docSet{}
	for id, _ := range coll.docs {
		matching[id] = struct{}{}
	}
	return matching
}

func (coll *Collection) find(field string, cmp func(string) bool) docSet {
	// resolve the field
	field = strings.ToLower(field)

	sf, ok := coll.docType.FieldByNameFunc(func(name string) bool {
		return strings.ToLower(name) == field
	})
	if !ok {
		panic("couldn't resolve field " + field)
	}

	matching := docSet{}

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

func (coll *Collection) findRange(field, start, end string) docSet {

	start = strings.ToLower(start)
	end = strings.ToLower(end)
	return coll.find(field, func(foo string) bool {
		foo = strings.ToLower(foo)
		return foo >= start && foo <= end
	})
}
func (coll *Collection) findExact(field, val string) docSet {
	val = strings.ToLower(val)
	return coll.find(field, func(foo string) bool {
		foo = strings.ToLower(foo)
		return foo == val
	})
}

func (coll *Collection) findContains(field, val string) docSet {
	val = strings.ToLower(val)
	return coll.find(field, func(foo string) bool {
		foo = strings.ToLower(foo)
		return strings.Contains(foo, val)
	})
}

func Read(in io.Reader, referenceDoc interface{}) (*Collection, error) {
	var err error

	cookie := []byte{0, 0, 0, 0}
	_, err = io.ReadFull(in, cookie)
	if err != nil {
		return nil, err
	}
	for i := 0; i < len(magicCookie); i++ {
		if cookie[i] != magicCookie[i] {
			return nil, fmt.Errorf("unrecognised file format")
		}
	}

	dec := gob.NewDecoder(in)

	coll := NewCollection(referenceDoc)

	var count int
	err = dec.Decode(&count)
	if err != nil {
		return nil, err
	}

	//inType := reflect.PtrTo(coll.docType)

	for i := 0; i < count; i++ {
		var key string
		doc := reflect.New(coll.docType)
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

	return coll, err
}

func (coll *Collection) Write(out io.Writer) error {
	var err error

	_, err = out.Write(magicCookie)
	if err != nil {
		return err
	}

	enc := gob.NewEncoder(out)

	err = enc.Encode(len(coll.docs))
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

// Find executes a query and fills out a slice containing the results.
// result must be a pointer to a slice
// eg
// var out []*Document
// coll.Find(q, &out)
func (coll *Collection) Find(q *Query, result interface{}) {
	var resultv, slicev reflect.Value
	var elementt reflect.Type
	var typeOK = false
	// we're very picky about what we shove the results into...
	resultv = reflect.ValueOf(result)
	if resultv.Kind() == reflect.Ptr {
		slicev = resultv.Elem()

		if slicev.Kind() == reflect.Slice {
			elementt = slicev.Type().Elem()
			if elementt.Kind() == reflect.Ptr {
				typeOK = true
			}
		}
	}
	if !typeOK {
		panic("result must be pointer to a slice of pointers")
	}

	ids := q.perform(coll)

	outv := reflect.MakeSlice(reflect.SliceOf(elementt), len(ids), len(ids))
	idx := 0
	for id, _ := range ids {
		doc := coll.docs[id]
		docv := reflect.ValueOf(doc)
		outv.Index(idx).Set(docv)
		idx++
	}
	resultv.Elem().Set(outv)
}

//
func (coll *Collection) Update(q *Query, modifyFn func(interface{})) int {
	ids := q.perform(coll)
	cnt := 0
	for id, _ := range ids {
		doc := coll.docs[id]
		modifyFn(doc)
		cnt++
	}
	return cnt
}
