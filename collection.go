package badger

import (
	"encoding/gob"
	"fmt"
	"io"
	"os"
	"reflect"
	"strings"
	"sync"
	"time"
)

var magicCookie = []byte{'b', 'a', 'd', '2'}

// Collection holds a set of documents, all of the same type.
// Caveats:
// - have to store ptrs to structs
// - can only query on string and []string fields (but can store anything)
//
type Collection struct {
	sync.RWMutex
	docs         map[uintptr]interface{}
	docType      reflect.Type
	DefaultField string // field to search by default (mainly for the benefit of the query parser)
	filename     string // filename for saving
	dirty        bool
}

// NewCollection initialises a collection for holding documents of
// same type as referenceDoc.
// The contents of referenceDoc are unimportant - a zero object is
// fine. Only it's type is used.
func NewCollection(referenceDoc interface{}) *Collection {
	coll := &Collection{
		docs:    make(map[uintptr]interface{}),
		docType: reflect.TypeOf(referenceDoc),
	}

	if coll.docType.Kind() != reflect.Ptr {
		panic("doctype must be ptr")
	}
	if coll.docType.Elem().Kind() != reflect.Struct {
		panic("doctype must be ptr to struct")
	}

	return coll
}

func (coll *Collection) EnableAutosave(filename string) {
	// TODO: use a channel to receive an exit signal. Add a Close() method to Collection to send it.
	coll.filename = filename
	if coll.filename == "" {
		panic("no filename")
	}
	go func() {
		var interval time.Duration = 5 * time.Second
		for {
			time.Sleep(interval)
			if coll.dirty {
				fmt.Fprintf(os.Stderr, "Autosaving...\n")
				err := coll.Save(coll.filename)
				if err != nil {
					fmt.Fprintf(os.Stderr, "Autosave failed: %s\n", err)
					interval *= 2 // back off
				} else {
					fmt.Fprintf(os.Stderr, "Autosave OK\n")
					interval = 5 * time.Second
				}
			}
		}
	}()
}

func (coll *Collection) Save(filename string) error {
	tmpFilename := filename + ".new"
	outFile, err := os.Create(tmpFilename)
	if err != nil {
		return err
	}
	coll.Write(outFile)
	outFile.Close()
	err = os.Rename(tmpFilename, filename)
	if err != nil {
		return err
	}
	return nil
}

func (coll *Collection) Count() int {
	coll.RLock()
	defer coll.RUnlock()
	return len(coll.docs)
}

// ValidField returns a list of valid field names
func (coll *Collection) ValidFields() []string {
	coll.RLock()
	defer coll.RUnlock()
	return validFields(coll.docType.Elem())
}

// TODO: update to:
// 1) handler pointer members
// 2) filter out unwanted members (eg functions)
func validFields(typ reflect.Type) []string {
	fields := []string{}
	for i := 0; i < typ.NumField(); i++ {
		sf := typ.Field(i)
		if sf.Type.Kind() == reflect.Struct {
			childFields := validFields(sf.Type)
			if !sf.Anonymous {
				for j, _ := range childFields {
					childFields[j] = sf.Name + "." + childFields[j]
				}
			}
			fields = append(fields, childFields...)
		} else {
			fields = append(fields, sf.Name)
		}
	}
	return fields
}

func (coll *Collection) Put(doc interface{}) {
	t := reflect.TypeOf(doc)
	if t != coll.docType {
		panic(fmt.Sprintf("doc type mismatch (got %s, expecting %s)", t, coll.docType))
	}
	key := reflect.ValueOf(doc).Pointer()

	coll.Lock()
	defer coll.Unlock()

	coll.docs[key] = doc
	coll.dirty = true
}

/*
func (coll *Collection) Get(id string) interface{} {
	coll.RLock()
	defer coll.RUnlock()
	return coll.docs[id]
}
*/

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

	sf, ok := coll.docType.Elem().FieldByNameFunc(func(name string) bool {
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
	coll.Lock()
	defer coll.Unlock()

	var count int
	err = dec.Decode(&count)
	if err != nil {
		return nil, err
	}

	//inType := reflect.PtrTo(coll.docType)

	for i := 0; i < count; i++ {
		doc := reflect.New(coll.docType.Elem())
		err = dec.Decode(doc.Interface())
		if err != nil {
			return nil, err
		}
		key := doc.Pointer()
		coll.docs[key] = doc.Interface()
	}

	return coll, err
}

func (coll *Collection) Write(out io.Writer) error {
	var err error

	coll.RLock()
	defer coll.RUnlock()
	_, err = out.Write(magicCookie)
	if err != nil {
		return err
	}

	enc := gob.NewEncoder(out)

	err = enc.Encode(len(coll.docs))
	if err != nil {
		return err
	}
	for _, doc := range coll.docs {
		err = enc.Encode(doc)
		if err != nil {
			return err
		}
	}
	coll.dirty = false
	return nil
}

// Find executes a query and fills out a slice containing the results.
// result must be a pointer to a slice
// eg
// var out []*Document
// coll.Find(q, &out)
func (coll *Collection) Find(q Query, result interface{}) {
	coll.RLock()
	defer coll.RUnlock()
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
func (coll *Collection) Update(q Query, modifyFn func(interface{})) int {
	coll.Lock()
	defer coll.Unlock()
	ids := q.perform(coll)
	cnt := 0
	for id, _ := range ids {
		doc := coll.docs[id]
		modifyFn(doc)
		cnt++
	}
	coll.dirty = true
	return cnt
}
