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
	if os.Args[1] == "stories" && len(os.Args) == 3 {
		err = SaveStories(os.Args[2])
	} else if os.Args[1] == "scrape" && len(os.Args) == 4 {
		err = Scrape(os.Args[2], os.Args[3])
	} else if os.Args[1] == "train" && len(os.Args) == 5 {
		err = errors.New("not yet implemented")
	} else if os.Args[1] == "scoresabove" && len(os.Args) == 4 {
		err = ScoresAbove(os.Args[2], os.Args[3])
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
		`Usage: hn-ranker stories <output.json>
       hn-ranker scrape <input.json> <output-dir>
       hn-ranker train <list.json> <post-dir> <classifier-out.json>
       hn-ranker scoresabove <list.json> <score>`)
	os.Exit(1)
}
