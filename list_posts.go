package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
)

func ListPosts(output string) error {
	newList, newErrs := FetchPosts(NewStories)
	topList, topErrs := FetchPosts(TopStories)

	var newPosts []*Post
	var topPosts []*Post

	log.Print("Listing new posts...")

	for newPost := range newList {
		newPosts = append(newPosts, newPost)
	}

	log.Print("Listing top posts...")

	for topPost := range topList {
		topPosts = append(topPosts, topPost)
	}

	if err := <-newErrs; err != nil {
		log.Println("Stopped fetching new posts due to error:", err)
	}
	if err := <-topErrs; err != nil {
		log.Println("Stopped fetching top posts due to error:", err)
	}

	jsonObj := map[string]interface{}{
		"new": newPosts,
		"top": topPosts,
	}

	data, err := json.Marshal(jsonObj)
	if err != nil {
		return err
	}

	return ioutil.WriteFile(output, data, 0755)
}
