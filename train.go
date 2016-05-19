package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/unixpickle/hn-ranker/hnclass"
)

var OutputScoreCutoffs = []int{2, 5, 10, 50}

const (
	ClassifierNameEnvVar = "HN_CLASSIFIER"
)

func Train(storyListFile, postDump, classifierOut string) error {
	log.Println("Parsing story list...")

	storyFile, err := ioutil.ReadFile(storyListFile)
	if err != nil {
		return err
	}

	var stories []*StoryItem
	if err := json.Unmarshal(storyFile, &stories); err != nil {
		return err
	}

	log.Println("Reading story data...")
	storyData, scores := loadStoryData(stories, postDump)

	log.Println("Creating feature map...")
	features := hnclass.NewFeatureMap(storyData)
	log.Printf("Feature counts: %d content, %d title, %d hostname",
		len(features.ContentKeywords), len(features.TitleKeywords), len(features.HostNames))

	log.Println("Initializing classifier...")
	classifier, err := makeClassifier(features, len(OutputScoreCutoffs)+1)
	if err != nil {
		return err
	}

	log.Println("Making feature vectors...")
	vecs := makeFeatureVectors(storyData, features)

	log.Println("Training...")
	classifier.Train(vecs, makeClasses(scores))

	log.Println("Saving classifier...")
	data := hnclass.Serialize(classifier, features)
	return ioutil.WriteFile(classifierOut, data, 0755)
}

func loadStoryData(stories []*StoryItem, postDump string) (data []*hnclass.StoryData,
	scores []int) {
	for _, story := range stories {
		fileName := strconv.FormatInt(story.ID, 10) + ".txt"
		postFile := filepath.Join(postDump, fileName)
		contents, err := ioutil.ReadFile(postFile)
		if err != nil {
			continue
		}

		var hostString string
		if parsedURL, _ := url.Parse(story.URL); parsedURL != nil {
			hostString = parsedURL.Host
		}

		storyData := &hnclass.StoryData{
			Title:    story.Title,
			Content:  string(contents),
			HostName: hostString,
			Time:     time.Unix(story.Time, 0),
		}
		data = append(data, storyData)
		scores = append(scores, story.Score)
	}
	return
}

func makeClassifier(features *hnclass.FeatureMap,
	classCount int) (hnclass.TrainableClassifier, error) {
	classifierName := os.Getenv(ClassifierNameEnvVar)
	if classifierName == "" {
		return nil, fmt.Errorf("missing %s environment variable", ClassifierNameEnvVar)
	}
	maker, ok := hnclass.ClassifierMakers[classifierName]
	if !ok {
		return nil, fmt.Errorf("invalid classifier name: %s", classifierName)
	}
	classifier, err := maker(features, classCount)
	if err != nil {
		return nil, err
	}
	return classifier, nil
}

func makeFeatureVectors(data []*hnclass.StoryData, m *hnclass.FeatureMap) []hnclass.FeatureVector {
	res := make([]hnclass.FeatureVector, len(data))
	for i, s := range data {
		res[i] = hnclass.NewFeatureVector(s, m)
	}
	return res
}

func makeClasses(scores []int) []int {
	classes := make([]int, len(scores))
	for i, score := range scores {
		var class int
		for _, c := range OutputScoreCutoffs {
			if score >= c {
				class++
			}
		}
		classes[i] = class
	}
	return classes
}
