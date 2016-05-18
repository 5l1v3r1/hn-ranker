# hn-ranker

This is an attempt to use Machine Learning to predict the scores which [Hacker News](https://news.ycombinator.com) articles will receive.

# Usage

First things first, install [Go](https://golang.org) and [setup a GOPATH](https://golang.org/doc/code.html#Workspaces). Next, download this code as follows:

```
$ go get github.com/unixpickle/hn-ranker
```

Now, change to the source directory so you can run the code:

```
$ cd $GOPATH/src/github.com/unixpickle/hn-ranker
```

In order to get a working classifier, you must go through several steps. First, fetch metadata about a large number of stories:

```
$ go run *.go stories ./story_metadata.json
2016/05/18 16:45:56 Fetching... (press ctrl+C to finish).
2016/05/18 16:46:01 Gotten story from 72 hours ago (1 stories)
2016/05/18 16:46:01 Gotten story from 72 hours ago (2 stories)
2016/05/18 16:46:01 Gotten story from 72 hours ago (3 stories)
2016/05/18 16:46:02 Gotten story from 72 hours ago (4 stories)
...
```

The above command fetches story metadata and will save it in `story_metadata.json` once you terminate the program with Control+C or once an error is encountered.

Next, you must fetch the actual contents of every story. This involves scraping many linked URLs (one per story). You can do this as follows:

```
$ go run *.go scrape ./story_metadata.json story_contents/
```

This will create a `story_contents` directory with a list of `.txt` files. While scraping, the `scrape` sub-command may log many errors to the console. While the `scrape` sub-command does log any errors it encounters, it continues fetching stories in spite of these errors. This prevents stories with broken links from holding up the entire data mining process.

**TODO:** document how to train some kind of classifier with the mined data. In order to add this, I will first have to figure out *how it will work*.
