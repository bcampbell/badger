package query

import (
	//	"fmt"
	"testing"
)

// just check they parse without syntax errors
func TestBasic(t *testing.T) {

	testQueries := []string{
		"",
		`tags:citrus`,
		`tags:"wibble pibble foo" headline: wibble`,
		`urls:/sport/`,
		`publication.domain:"www.dailymail.co.uk"`,
		`site:"www.dailymail.co.uk"`,
		`stuff AND nonsense OR grapefruit`,
		`(stuff AND nonsense) OR grapefruit`,
		`stuff AND (nonsense OR grapefruit)`,
		`published:[ 2010-02-01 TO 2010-04-15]`,
		`published:[ TO 2010-04-15]`,
		`published:[ 2010-04-15 TO]`,
		//`headline:(citrus -grapefruit)`,
		//`published: ..2010-04-15`,

		"(fred bloggs) OR (bob smith)",
	}

	for _, qs := range testQueries {
		_, err := Parse(qs, "title")
		if err != nil {
			t.Errorf(`Parse(%s) failed: %s`, qs, err)
			continue
		}

		//		fmt.Println(qs)
		//		fmt.Println(DumpTree(q, 0))

	}
}
