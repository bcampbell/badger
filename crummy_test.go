package crummy

import (
	//	"fmt"
	"testing"
)

type TestDoc struct {
	ID     string
	Colour string
	Tags   []string
}

func TestSimple(t *testing.T) {

	store := NewStore(TestDoc{})

	testData := []*TestDoc{
		&TestDoc{"1", "red", []string{"primary", "reddish"}},
		&TestDoc{"two", "green", []string{"primary"}},
		&TestDoc{"3", "blue", []string{"primary"}},
		&TestDoc{"4", "pink", []string{"reddish"}},
		&TestDoc{"five", "crimson", []string{"reddish"}},
	}
	for _, dat := range testData {
		store.Put(dat.ID, dat)
	}

	if store.Count() != 5 {
		t.Error("Count error")
	}

	greens := store.FindExact("Colour", "green")
	//	fmt.Println(greens)
	reds := store.FindExact("Colour", "crimson")
	reds = Union(reds, store.FindExact("Colour", "pink"))
	reds = Union(reds, store.FindExact("Colour", "red"))

	//	fmt.Println(reds)
	if len(greens) != 1 {
		t.Error("Wrong number of greens")
	}
	if len(reds) != 3 {
		t.Error("Wrong number of reds")
	}

	//
	if len(store.FindExact("Tags", "reddish")) != 3 {
		t.Error("wrong number tagged reddish")
	}
	if len(store.FindExact("Tags", "uber")) != 0 {
		t.Error("wrong number tagged uber")
	}

	if len(store.FindContaining("Tags", "reddish")) != 3 {
		t.Error("wrong number tagged reddish")
	}
}
