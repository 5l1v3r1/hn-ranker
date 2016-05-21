# Abstract

The goal of this project was to predict the scores of [Hacker News](https://news.ycombinator.com) articles. To do this, I tried training a feedforward ANN on two week's worth of Hacker News posts. During preprocessing, articles were decomposed into a list of keyword frequencies which was then fed to the network.

# Results

**Brief summary:** the results were not very impressive. The ANN learned to classify the documents it was trained on, but it did not generalize well to other articles (i.e. it overfit the data).

| Data set | Total | 0-1 | 2-4 | 5-9 | 10-49 | 50+ |
|----------|-------|-----|-----|-----|-------|-----|
| Training | 8866/10194 | 4294/4618 | 3455/3810 | 413/675 | 269/485 | 435/606 |
| Cross    | 2007/4368  | 1126/1980 | 838/1633  | 14/269  | 11/230  | 18/256  |


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
