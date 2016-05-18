package main

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"strconv"
	"time"
)

const (
	storyType = "story"
	apiRoot   = "https://hacker-news.firebaseio.com/v0/"
)

type Item struct {
	Time int64  `json:"time"`
	Type string `json:"type"`
	ID   int64  `json:"id"`
}

type StoryItem struct {
	Item
	Title string `json:"title"`
	URL   string `json:"url"`
	Score int    `json:"score"`
}

func FetchStoryItems(beforeTime time.Time) (<-chan *StoryItem, <-chan error) {
	storyChan := make(chan *StoryItem)
	errChan := make(chan error, 1)

	go func() {
		defer close(storyChan)
		defer close(errChan)

		var latestID int64
		if err := fetchAPIPage("maxitem.json", &latestID); err != nil {
			errChan <- err
			return
		}

		startID, err := firstItemBeforeTime(beforeTime, latestID)
		if err != nil {
			errChan <- err
			return
		}

		for id := startID; id >= 0; id-- {
			var s StoryItem
			if err := fetchItem(id, &s); err != nil {
				errChan <- err
				return
			}
			if s.Type != storyType {
				continue
			}
			storyChan <- &s
		}
	}()

	return storyChan, errChan
}

func firstItemBeforeTime(t time.Time, maxId int64) (int64, error) {
	upperBound := maxId
	var lowerBound int64

	subAmount := int64(1)
	for subAmount < maxId {
		id := maxId - subAmount
		var item Item
		if err := fetchItem(id, &item); err != nil {
			return 0, err
		}
		postTime := time.Unix(item.Time, 0)
		if postTime.Before(t) {
			lowerBound = id
			break
		}
		subAmount *= 2
	}

	if subAmount > maxId {
		return 0, nil
	}

	for upperBound > lowerBound+1 {
		midPoint := (upperBound + lowerBound) / 2
		var item Item
		if err := fetchItem(midPoint, &item); err != nil {
			return 0, err
		}
		postTime := time.Unix(item.Time, 0)
		if postTime.Before(t) {
			lowerBound = midPoint
		} else {
			upperBound = midPoint
		}
	}

	return lowerBound, nil
}

func fetchItem(id int64, obj interface{}) error {
	idStr := "item/" + strconv.FormatInt(id, 10) + ".json"
	return fetchAPIPage(idStr, obj)
}

func fetchAPIPage(path string, obj interface{}) error {
	u := apiRoot + path
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
