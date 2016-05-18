package main

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"log"
	"strconv"
)

func ScoresAbove(listFile string, threshold string) error {
	thresholdNum, err := strconv.Atoi(threshold)
	if err != nil {
		return errors.New("invalid threshold: " + threshold)
	}

	data, err := ioutil.ReadFile(listFile)
	if err != nil {
		return err
	}

	var list []*StoryItem
	if err := json.Unmarshal(data, &list); err != nil {
		return err
	}

	var count, total int
	for _, s := range list {
		if s.Score > thresholdNum {
			count++
		}
		total++
	}

	log.Printf("Matched %d/%d (%0.2f%%)", count, total, float64(count)/float64(total))
	return nil
}
