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

	"github.com/unixpickle/weakai/neuralnet"
)

var outputScoreCutoffs = []int{2, 5, 10, 50}

const maxEpochs = 100000

func Train(storyListFile, postDump, classifierOut string) error {
	config, err := getTrainConfig()
	if err != nil {
		return err
	}

	log.Println("Parsing story list...")

	storyFile, err := ioutil.ReadFile(storyListFile)
	if err != nil {
		return err
	}

	var stories []*StoryItem
	if err := json.Unmarshal(storyFile, &stories); err != nil {
		return err
	}

	log.Println("Generating features and reading samples...")

	samples, features := samplesAndFeatures(stories, postDump)
	featureVecs := make([][]float64, len(samples))
	for i, s := range samples {
		featureVecs[i] = s.FeatureVector(features)
	}

	network, trainer := networkAndTrainer(samples, featureVecs, config)

	log.Println("Training...")

	trainer.TrainInteractive(network)

	classifier := &Classifier{
		Features: features,
		Network:  network,
	}

	data := classifier.Encode()

	return ioutil.WriteFile(classifierOut, data, 0755)
}

func samplesAndFeatures(stories []*StoryItem, postDump string) ([]*Sample, *Features) {
	seenContentKeywords := map[string]bool{}
	seenTitleKeywords := map[string]bool{}
	seenHostNames := map[string]bool{}

	samples := make([]*Sample, 0, len(stories))
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
			seenHostNames[hostString] = true
		}
		time := time.Unix(story.Time, 0)
		sample := NewSample(story.Title, string(contents), hostString, time)
		sample.Score = story.Score
		samples = append(samples, sample)

		for key := range extractKeywords(string(contents)) {
			seenContentKeywords[key] = true
		}
		for key := range extractKeywords(story.Title) {
			seenTitleKeywords[key] = true
		}
	}

	contentKeywords := make([]string, 0, len(seenContentKeywords))
	titleKeywords := make([]string, 0, len(seenTitleKeywords))
	hostNames := make([]string, 0, len(seenHostNames))

	for key := range seenContentKeywords {
		contentKeywords = append(contentKeywords, key)
	}
	for key := range seenTitleKeywords {
		titleKeywords = append(titleKeywords, key)
	}
	for key := range seenHostNames {
		hostNames = append(hostNames, key)
	}

	features := &Features{
		TitleKeywords:   titleKeywords,
		ContentKeywords: contentKeywords,
		HostNames:       hostNames,
	}

	return samples, features
}

func networkAndTrainer(samples []*Sample, vecs [][]float64,
	config *trainConfig) (*neuralnet.Network, *neuralnet.SGD) {
	network, err := neuralnet.NewNetwork([]neuralnet.LayerPrototype{
		&neuralnet.DenseParams{
			Activation:  neuralnet.Sigmoid{},
			InputCount:  len(vecs[0]),
			OutputCount: config.HiddenCount,
		},
		&neuralnet.DenseParams{
			Activation:  neuralnet.Sigmoid{},
			InputCount:  config.HiddenCount,
			OutputCount: len(outputScoreCutoffs) + 1,
		},
	})
	if err != nil {
		fmt.Fprintln(os.Stderr, "Could not create network:", err)
		os.Exit(1)
	}

	outputs := make([][]float64, len(vecs))
	for i, sample := range samples {
		var idx int
		for idx < len(outputScoreCutoffs) && sample.Score > outputScoreCutoffs[idx] {
			idx++
		}
		outputs[i] = make([]float64, len(outputScoreCutoffs)+1)
		outputs[i][idx] = 1
	}

	sgd := &neuralnet.SGD{
		CostFunc: neuralnet.MeanSquaredCost{},
		Inputs:   vecs,
		Outputs:  outputs,
		StepSize: config.StepSize,
		Epochs:   maxEpochs,
	}

	return network, sgd
}

type trainConfig struct {
	HiddenCount int
	StepSize    float64
}

func getTrainConfig() (*trainConfig, error) {
	count, err := strconv.Atoi(os.Getenv("NEURALNET_HIDDEN_COUNT"))
	if err != nil {
		return nil, errors.New("missing NEURALNET_HIDDEN_COUNT env var")
	}
	stepSize, err := strconv.ParseFloat(os.Getenv("NEURALNET_STEP_SIZE"), 64)
	if err != nil {
		return nil, errors.New("missing NEURALNET_STEP_SIZE env var")
	}
	return &trainConfig{
		HiddenCount: count,
		StepSize:    stepSize,
	}, nil
}
