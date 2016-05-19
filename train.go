package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"time"
)

var outputScoreCutoffs = []int{2, 5, 10, 50}

const (
	maxEpochs       = 100000
	keywordUbiquity = 2
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
	stories = stories[:100]

	log.Println("Reading story data...")
	storyData, scores := loadStoryData(stories, postDump)

	log.Println("Creating feature map...")
	features := makeFeatureMap(storyData)
	log.Printf("Counts: %d content, %d title, %d hostname",
		len(features.ContentKeywords), len(features.TitleKeywords), len(features.HostNames))

	log.Println("Initializing classifier...")
	classifier, err := makeClassifier(features)
	if err != nil {
		return err
	}

	log.Println("Making feature vectors...")
	vecs := makeFeatureVectors(storyData, features)

	log.Println("Training...")
	classifier.Train(vecs, makeClasses(scores))

	log.Println("Saving classifier...")
	// TODO: encode classifier and feature map and
	// save it all to one file.
	//data := classifier.Serialize()
	//return ioutil.WriteFile(classifierOut, data, 0755)
	return errors.New("not yet implemented")
}

func loadStoryData(stories []*StoryItem, postDump string) (data []*StoryData, scores []int) {
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

		storyData := &StoryData{
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

func makeFeatureMap(data []*StoryData) *FeatureMap {
	seenContentKeywords := map[string]int{}
	seenTitleKeywords := map[string]bool{}
	seenHostNames := map[string]bool{}

	for _, storyData := range data {
		seenHostNames[storyData.HostName] = true
		for keyword := range extractKeywords(storyData.Content) {
			seenContentKeywords[keyword]++
		}
		for keyword := range extractKeywords(storyData.Title) {
			seenTitleKeywords[keyword] = true
		}
	}

	contentKeywords := make([]string, 0, len(seenContentKeywords))
	titleKeywords := make([]string, 0, len(seenTitleKeywords))
	hostNames := make([]string, 0, len(seenHostNames))

	for key, count := range seenContentKeywords {
		if count >= keywordUbiquity {
			contentKeywords = append(contentKeywords, key)
		}
	}
	for key := range seenTitleKeywords {
		titleKeywords = append(titleKeywords, key)
	}
	for key := range seenHostNames {
		hostNames = append(hostNames, key)
	}

	return &FeatureMap{
		TitleKeywords:   titleKeywords,
		ContentKeywords: contentKeywords,
		HostNames:       hostNames,
	}
}

func makeClassifier(features *FeatureMap) (TrainableClassifier, error) {
	classifierName := os.Getenv(ClassifierNameEnvVar)
	if classifierName == "" {
		return nil, fmt.Errorf("missing %s environment variable", ClassifierNameEnvVar)
	}
	maker, ok := ClassifierMakers[classifierName]
	if !ok {
		return nil, fmt.Errorf("invalid classifier name: %s", classifierName)
	}
	classifier, err := maker(features)
	if err != nil {
		return nil, err
	}
	return classifier, nil
}

func makeFeatureVectors(data []*StoryData, m *FeatureMap) []FeatureVector {
	res := make([]FeatureVector, len(data))
	for i, s := range data {
		res[i] = NewFeatureVector(s, m)
	}
	return res
}

func makeClasses(scores []int) []int {
	classes := make([]int, len(scores))
	for i, score := range scores {
		var class int
		for _, c := range outputScoreCutoffs {
			if score >= c {
				class++
			}
		}
		classes[i] = class
	}
	return classes
}
