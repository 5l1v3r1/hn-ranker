package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"net/url"
	"os"
	"os/signal"
	"path/filepath"
	"strconv"
	"time"

	"github.com/unixpickle/weakai/neuralnet"
)

var outputScoreCutoffs = []int{2, 5, 10, 50}

const (
	maxEpochs       = 100000
	keywordUbiquity = 2
)

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

	stories = stories[:100]

	log.Println("Reading story data...")
	storyData, scores := loadStoryData(stories, postDump)

	log.Println("Creating feature map...")
	features := makeFeatureMap(storyData)
	log.Printf("Counts: %d content, %d title, %d hostname",
		len(features.ContentKeywords), len(features.TitleKeywords), len(features.HostNames))

	log.Println("Making feature vectors...")
	vecs := makeFeatureVectors(storyData, features)

	log.Println("Making network...")
	network := makeNetwork(features, config)

	log.Println("Training...")

	killChan := make(chan struct{})
	go func() {
		c := make(chan os.Signal, 1)
		signal.Notify(c, os.Interrupt)
		<-c
		signal.Stop(c)
		fmt.Println("\nCaught interrupt. Ctrl+C again to terminate.")
		close(killChan)
	}()

	trainNetwork(network, vecs, scores, config, killChan)

	log.Println("Saving classifier...")

	classifier := &Classifier{
		Features: features,
		Network:  network,
	}

	data := classifier.Encode()

	return ioutil.WriteFile(classifierOut, data, 0755)
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

func makeNetwork(m *FeatureMap, c *trainConfig) *neuralnet.Network {
	network, err := neuralnet.NewNetwork([]neuralnet.LayerPrototype{
		&neuralnet.DenseParams{
			Activation:  neuralnet.Sigmoid{},
			InputCount:  m.VectorSize(),
			OutputCount: c.HiddenCount,
		},
		&neuralnet.DenseParams{
			Activation:  neuralnet.Sigmoid{},
			InputCount:  c.HiddenCount,
			OutputCount: len(outputScoreCutoffs) + 1,
		},
	})
	if err != nil {
		fmt.Fprintln(os.Stderr, "Could not create network:", err)
		os.Exit(1)
	}
	network.SetInput(make([]float64, m.VectorSize()))
	network.SetDownstreamGradient(make([]float64, len(network.Output())))
	return network
}

func makeFeatureVectors(data []*StoryData, m *FeatureMap) []FeatureVector {
	res := make([]FeatureVector, len(data))
	for i, s := range data {
		res[i] = NewFeatureVector(s, m)
	}
	return res
}

func trainNetwork(n *neuralnet.Network, f []FeatureVector, scores []int, c *trainConfig,
	cancel <-chan struct{}) {
	n.Randomize()
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
	for {
		rightCount := classifierNumRight(n, f, classes)
		log.Printf("Getting %d out of %d", rightCount, len(classes))
		perm := rand.Perm(len(f))
		for _, x := range perm {
			story := f[x]
			class := classes[x]
			sgdStepStory(n, story, class, c)
			select {
			case <-cancel:
				return
			default:
			}
		}
	}
}

func sgdStepStory(n *neuralnet.Network, f FeatureVector, class int, c *trainConfig) {
	inputVec := n.Input()
	downstream := n.DownstreamGradient()

	for i := range inputVec {
		inputVec[i] = 0
	}
	for _, v := range f {
		inputVec[v.Index] = v.Value
	}

	n.PropagateForward()

	expected := make([]float64, len(downstream))
	expected[class] = 1

	costFunc := neuralnet.MeanSquaredCost{}
	costFunc.Deriv(n, expected, downstream)

	n.PropagateBackward(false)

	n.StepGradient(-c.StepSize)
}

func classifierNumRight(n *neuralnet.Network, f []FeatureVector, classes []int) int {
	var rightCount int
	for i, x := range f {
		inputVec := n.Input()
		for i := range inputVec {
			inputVec[i] = 0
		}
		for _, v := range x {
			inputVec[v.Index] = v.Value
		}
		n.PropagateForward()
		var outputClass int
		var maxOutput float64
		for j, out := range n.Output() {
			if out > maxOutput {
				maxOutput = out
				outputClass = j
			}
		}
		if outputClass == classes[i] {
			rightCount++
		}
	}
	return rightCount
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
