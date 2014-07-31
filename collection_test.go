package badger

import (
	"bytes"
	"fmt"
	"testing"
)

type SubDoc struct {
	Name     string
	ShoeSize int
}

type TestDoc struct {
	ID      string
	Colour  string
	Tags    []string
	Details SubDoc
}

func dummyCollection() *Collection {
	coll := NewCollection(&TestDoc{})

	testData := []*TestDoc{
		&TestDoc{"1", "red", []string{"primary", "reddish"}, SubDoc{}},
		&TestDoc{"two", "green", []string{"primary"}, SubDoc{"Bob", 42}},
		&TestDoc{"3", "blue", []string{"primary"}, SubDoc{}},
		&TestDoc{"4", "pink", []string{"reddish"}, SubDoc{}},
		&TestDoc{"five", "crimson", []string{"reddish"}, SubDoc{}},
	}
	for _, dat := range testData {
		coll.Put(dat)
	}
	return coll
}

func TestFind(t *testing.T) {
	coll := dummyCollection()

	var out []*TestDoc

	q := NewContainsQuery("Colour", "e")
	coll.Find(q, &out)

	if len(out) != 3 {
		t.Error("Count error")
	}
}

func TestSimple(t *testing.T) {

	coll := dummyCollection()
	if coll.Count() != 5 {
		t.Error("Count error")
	}

	greens := coll.findExact("Colour", "green")
	//	fmt.Println(greens)
	reds := coll.findExact("Colour", "crimson")
	reds = Union(reds, coll.findExact("Colour", "pink"))
	reds = Union(reds, coll.findExact("Colour", "red"))

	//	fmt.Println(reds)
	if len(greens) != 1 {
		t.Error("Wrong number of greens")
	}
	if len(reds) != 3 {
		t.Error("Wrong number of reds")
	}

	//
	if len(coll.findExact("Tags", "reddish")) != 3 {
		t.Error("wrong number tagged reddish")
	}
	if len(coll.findExact("Tags", "uber")) != 0 {
		t.Error("wrong number tagged uber")
	}

	if len(coll.findContains("Tags", "reddish")) != 3 {
		t.Error("wrong number tagged reddish")
	}
}

func TestReadWrite(t *testing.T) {
	// save out the dummy data then read it back in

	coll := dummyCollection()
	var buf bytes.Buffer
	err := coll.Write(&buf)
	if err != nil {
		t.Error("write failed")
		return
	}

	coll2, err := Read(&buf, &TestDoc{})
	if err != nil {
		t.Errorf("read failed (%s)", err)
		return
	}

	if coll.Count() != coll2.Count() {
		t.Error("count mismatched")

	}
	// make sure result is kind of sane
	greens := coll2.findExact("Colour", "green")
	if len(greens) != 1 {
		t.Error("Wrong number of greens")
	}
	/*
		for id, _ := range coll.findAll() {
			a := coll.Get(id).(*TestDoc)

			b := coll2.Get(id).(*TestDoc)
			fmt.Printf("%v %v\n", a, b)
		}
	*/
}

func TestUpdate(t *testing.T) {
	coll := dummyCollection()
	visited := coll.Update(NewAllQuery(), func(a interface{}) {
		_ = a.(*TestDoc)
		//fmt.Printf("%v\n", doc)
	})

	if visited != coll.Count() {
		t.Error("Wrong number visited")
	}
}

func ExampleTestValidFields() {

	// test struct with anonymous embedded struct
	type ExtendedDoc struct {
		TestDoc
		Extra string
	}

	coll := NewCollection(&ExtendedDoc{})
	fields := coll.ValidFields()
	fmt.Printf("%v\n", fields)
	// Output:
	// [ID Colour Tags Details.Name Details.ShoeSize Extra]
}
