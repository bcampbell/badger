package main

// dump out parsed tree of query passed in on commandline

import (
	"flag"
	"fmt"
	"os"
	"semprini/badger/query"
)

func main() {
	flag.Parse()

	qs := flag.Arg(0)
	fmt.Println(qs)

	q, err := query.Parse(qs, "default")
	if err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: %s\n", err)
		os.Exit(1)
	}
	fmt.Println(query.DumpTree(q, 0))
}
