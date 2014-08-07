package query

import (
	//	"fmt"
	"github.com/bcampbell/badger"
	"strings"
	"testing"
)

type TestDoc struct {
	ID    string
	Title string
	Date  string
	Tags  []string
}

func TestQueries(t *testing.T) {

	testDocs := []*TestDoc{
		&TestDoc{ID: "1", Title: "Moon made of cheese", Tags: []string{"moon", "cheese"}, Date: "2010-06-14"},
		&TestDoc{ID: "2", Title: "Weekly Citrus Roundup", Tags: []string{"citrus"}, Date: "2010-06-14T10:20"},
		&TestDoc{ID: "3", Title: "Recipe: Zesty Cheese", Tags: []string{"cheese", "lemon", "citrus"}, Date: ""},
		&TestDoc{ID: "4", Title: "Grapefruit is the New Lemon", Tags: []string{"citrus"}, Date: "T11:52"},
		&TestDoc{ID: "5", Title: "De la terre à la lune", Tags: []string{"moon"}, Date: "1865-01-01"},
	}

	tests := []struct{ q, expect string }{
		{"moon", "1"},
		{"tags:(cheese OR moon)", "1,3,5"},
		{"tags:(cheese moon)", "1"}, // implict AND
		{"tags:(cheese AND moon)", "1"},
		{"date:[2010-06-14 TO]", "1,2"},   // >=
		{"date:[TO 2010-06-14]", "1,2,5"}, // <=
	}

	coll := badger.NewCollection(&TestDoc{})
	for _, doc := range testDocs {
		coll.Put(doc)
		//fmt.Println(doc)

	}
	//		fmt.Println(DumpTree(q, 0))

	for _, test := range tests {
		q, err := Parse(test.q, coll.ValidFields(), "title")
		if err != nil {
			panic(err)
		}
		var matches []*TestDoc
		coll.Find(q, &matches)
		ids := make([]string, 0)
		for _, doc := range matches {
			ids = append(ids, doc.ID)
		}
		got := strings.Join(ids, ",")
		if test.expect != got {
			t.Errorf(`Query error (%s) got %s, expected %s`, test.q, got, test.expect)
		}
	}

}
