package crummy

import (
	//	"fmt"
	"bytes"
	"testing"
)

type TestDoc struct {
	ID     string
	Colour string
	Tags   []string
}

func dummyCollection() *Collection {
	coll := NewCollection(TestDoc{})

	testData := []*TestDoc{
		&TestDoc{"1", "red", []string{"primary", "reddish"}},
		&TestDoc{"two", "green", []string{"primary"}},
		&TestDoc{"3", "blue", []string{"primary"}},
		&TestDoc{"4", "pink", []string{"reddish"}},
		&TestDoc{"five", "crimson", []string{"reddish"}},
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
	// TODO: run through and ensure docs are identical!
}
