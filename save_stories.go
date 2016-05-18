package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/signal"
	"time"
)

const minPostAge = time.Hour * 24 * 3

func SaveStories(output string) error {
	storyChan, errs := FetchStoryItems(time.Now().Add(-minPostAge))

	var stories []*StoryItem

	log.Println("Fetching... (press ctrl+C to finish).")

	terminateChan := make(chan struct{})
	go func() {
		c := make(chan os.Signal, 1)
		signal.Notify(c, os.Interrupt)
		<-c
		signal.Stop(c)
		fmt.Println("\nCaught interrupt. Ctrl+C again to terminate.")
		close(terminateChan)
	}()

StoryLoop:
	for {
		select {
		case <-terminateChan:
			break StoryLoop
		default:
		}

		select {
		case story := <-storyChan:
			diff := time.Now().Sub(time.Unix(story.Time, 0))
			stories = append(stories, story)
			log.Printf("Gotten story from %d hours ago (%d stories)", diff/time.Hour, len(stories))
		case <-terminateChan:
			break StoryLoop
		}
	}

	select {
	case err := <-errs:
		log.Println("Error while fetching:", err)
	default:
	}

	data, err := json.Marshal(stories)
	if err != nil {
		return err
	}

	return ioutil.WriteFile(output, data, 0755)
}
