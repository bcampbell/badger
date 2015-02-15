package badger

import (
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

	greens := NewExactQuery("Colour", "green").perform(coll)
	//	fmt.Println(greens)
	reds := NewExactQuery("Colour", "crimson").perform(coll)
	reds = Union(reds, NewExactQuery("Colour", "pink").perform(coll))
	reds = Union(reds, NewExactQuery("Colour", "red").perform(coll))

	//	fmt.Println(reds)
	if len(greens) != 1 {
		t.Error("Wrong number of greens")
	}
	if len(reds) != 3 {
		t.Error("Wrong number of reds")
	}

	//
	if len(NewExactQuery("Tags", "reddish").perform(coll)) != 3 {
		t.Error("wrong number tagged reddish")
	}
	if len(NewExactQuery("Tags", "uber").perform(coll)) != 0 {
		t.Error("wrong number tagged uber")
	}

	if len(NewContainsQuery("Tags", "reddish").perform(coll)) != 3 {
		t.Error("wrong number tagged reddish")
	}

	notgreens := NewNOTQuery(NewExactQuery("Colour", "green")).perform(coll)
	if len(notgreens) != 4 {
		t.Error("wrong number not green")
	}

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
