package crummy

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
	coll := NewCollection(TestDoc{})

	testData := []*TestDoc{
		&TestDoc{"1", "red", []string{"primary", "reddish"}, SubDoc{}},
		&TestDoc{"two", "green", []string{"primary"}, SubDoc{"Bob", 42}},
		&TestDoc{"3", "blue", []string{"primary"}, SubDoc{}},
		&TestDoc{"4", "pink", []string{"reddish"}, SubDoc{}},
		&TestDoc{"five", "crimson", []string{"reddish"}, SubDoc{}},
	}
	for _, dat := range testData {
		coll.Put(dat.ID, dat)
	}
	return coll
}

func TestSimple(t *testing.T) {

	coll := dummyCollection()
	if coll.Count() != 5 {
		t.Error("Count error")
	}

	greens := coll.FindExact("Colour", "green")
	//	fmt.Println(greens)
	reds := coll.FindExact("Colour", "crimson")
	reds = Union(reds, coll.FindExact("Colour", "pink"))
	reds = Union(reds, coll.FindExact("Colour", "red"))

	//	fmt.Println(reds)
	if len(greens) != 1 {
		t.Error("Wrong number of greens")
	}
	if len(reds) != 3 {
		t.Error("Wrong number of reds")
	}

	//
	if len(coll.FindExact("Tags", "reddish")) != 3 {
		t.Error("wrong number tagged reddish")
	}
	if len(coll.FindExact("Tags", "uber")) != 0 {
		t.Error("wrong number tagged uber")
	}

	if len(coll.FindContains("Tags", "reddish")) != 3 {
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
	}

	coll2, err := Read(&buf, TestDoc{})
	if err != nil {
		t.Error("read failed")
	}

	if coll.Count() != coll2.Count() {
		t.Error("count mismatched")

	}
	// make sure result is kind of sane
	greens := coll2.FindExact("Colour", "green")
	if len(greens) != 1 {
		t.Error("Wrong number of greens")
	}
	/*
		for id, _ := range coll.FindAll() {
			a := coll.Get(id).(*TestDoc)

			b := coll2.Get(id).(*TestDoc)
			fmt.Printf("%v %v\n", a, b)
		}
	*/

}
