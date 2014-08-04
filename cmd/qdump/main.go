package main

// dump out parsed tree of query passed in on commandline

import (
	"flag"
	"fmt"
	//	"github.com/bcampbell/badger"
	"github.com/bcampbell/badger/query"
	"os"
)

func main() {
	flag.Parse()

	qs := flag.Arg(0)
	fmt.Println(qs)

	q, err := query.Parse(qs, []string{"title", "author", "tags", "content", "published"}, "content")
	if err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: %s\n", err)
		os.Exit(1)
	}
	fmt.Println(q)
}
