package main

import (
	"errors"
	"fmt"
	"os"
)

func main() {
	if len(os.Args) < 2 {
		dieUsage()
	}

	var err error
	if os.Args[1] == "listposts" && len(os.Args) == 3 {
		err = ListPosts(os.Args[2])
	} else if os.Args[1] == "fetchposts" && len(os.Args) == 4 {
		err = errors.New("not yet implemented")
	} else if os.Args[1] == "train" && len(os.Args) == 4 {
		err = errors.New("not yet implemented")
	} else {
		dieUsage()
	}

	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func dieUsage() {
	fmt.Fprintln(os.Stderr,
		`Usage: hn-ranker listposts <output.json>
       hn-ranker fetchposts <input.json> <output-dir>
       hn-ranker train <post-dir> <trained.json>`)
	os.Exit(1)
}
