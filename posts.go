package main

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"strconv"
)

type Collection string

const (
	NewStories  Collection = "newstories"
	TopStories             = "topstories"
	BestStories            = "beststories"
)

const APIRoot = "https://hacker-news.firebaseio.com/v0/"

type Post struct {
	Title string `json:"title"`
	Time  int64  `json:"time"`
	URL   string `json:"url"`
}

func FetchPosts(c Collection) (<-chan *Post, <-chan error) {
	postChan := make(chan *Post)
	errChan := make(chan error)

	go func() {
		defer close(postChan)
		defer close(errChan)

		var postList []int64
		if err := fetchAPIPage(string(c)+".json", &postList); err != nil {
			errChan <- err
			return
		}

		for _, id := range postList {
			var p Post
			intStr := strconv.FormatInt(id, 10)
			if err := fetchAPIPage("item/"+intStr, &p); err != nil {
				errChan <- err
				return
			}
			postChan <- &p
		}
	}()

	return postChan, errChan
}

func fetchAPIPage(path string, obj interface{}) error {
	u := APIRoot + path
	resp, err := http.Get(u)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	return json.Unmarshal(data, obj)
}
