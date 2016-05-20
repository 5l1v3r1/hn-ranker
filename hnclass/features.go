package hnclass

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

const (
	defaultContentUbiquity = 2
	defaultTitleUbiquity   = 1
	defaultHostUbiquity    = 1
)

// StoryData contains the raw data of a story,
// before it is converter into a feature vector.
type StoryData struct {
	Title    string
	Content  string
	HostName string
	Time     time.Time
}

// A FeatureMap describes how to map data from
// a story to a vector of scalers.
type FeatureMap struct {
	TitleKeywords   []string
	ContentKeywords []string
	HostNames       []string

	Offset float64
	Scale  float64
}

// NewFeatureMap generates a FeatureMap which
// contains all of the keywords from all of the
// stories in a list.
func NewFeatureMap(stories []*StoryData) *FeatureMap {
	seenContentKeywords := map[string]int{}
	seenTitleKeywords := map[string]int{}
	seenHostNames := map[string]int{}

	for _, storyData := range stories {
		seenHostNames[storyData.HostName]++
		for keyword := range extractKeywords(storyData.Content) {
			seenContentKeywords[keyword]++
		}
		for keyword := range extractKeywords(storyData.Title) {
			seenTitleKeywords[keyword]++
		}
	}

	contentKeywords := make([]string, 0, len(seenContentKeywords))
	titleKeywords := make([]string, 0, len(seenTitleKeywords))
	hostNames := make([]string, 0, len(seenHostNames))

	ubiquities := []int{
		getUbiquity(ContentUbiquityEnvVar, defaultContentUbiquity),
		getUbiquity(TitleUbiquityEnvVar, defaultTitleUbiquity),
		getUbiquity(HostUbiquityEnvVar, defaultHostUbiquity),
	}
	counts := []map[string]int{seenContentKeywords, seenTitleKeywords, seenHostNames}
	slices := []*[]string{&contentKeywords, &titleKeywords, &hostNames}

	for i, ubiquity := range ubiquities {
		slice := slices[i]
		for word, count := range counts[i] {
			if count >= ubiquity {
				(*slice) = append(*slice, word)
			}
		}
	}

	return &FeatureMap{
		TitleKeywords:   titleKeywords,
		ContentKeywords: contentKeywords,
		HostNames:       hostNames,

		// Computed under the assumption that no keywords
		// were pruned, or at least that a small fraction
		// of them were.
		Offset: -2.5,
		Scale:  0.4,
	}
}

// VectorSize returns the total number of features in
// the feature vectors created for f.
// This includes space for date/time features.
func (f *FeatureMap) VectorSize() int {
	return len(f.TitleKeywords) + len(f.ContentKeywords) + len(f.HostNames) + 24 + 7
}

type FeatureValue struct {
	Index int
	Value float64
}

// A FeatureVector is a vector of scaler values.
// It is represented as an array of index-value
// pairs, which are sorted by index.
// This is done to allow sparsity: if an index is
// not present, its corresponding value is assumed
// to be 0.
type FeatureVector []FeatureValue

// NewFeatureVector generates a FeatureVector which
// represents the given story data for the given
// mapping of features, m.
func NewFeatureVector(data *StoryData, m *FeatureMap) FeatureVector {
	var res FeatureVector

	titleKeywords := extractKeywords(data.Title)
	contentKeywords := extractKeywords(data.Content)
	for i, x := range m.ContentKeywords {
		val, ok := contentKeywords[x]
		if ok {
			res = append(res, FeatureValue{i, val})
		}
	}
	startIdx := len(m.ContentKeywords)
	for i, x := range m.TitleKeywords {
		val, ok := titleKeywords[x]
		if ok {
			res = append(res, FeatureValue{startIdx + i, val})
		}
	}
	startIdx += len(m.TitleKeywords)
	for i, host := range m.HostNames {
		if host == data.HostName {
			res = append(res, FeatureValue{startIdx + i, 1})
			break
		}
	}

	startIdx += len(m.HostNames)

	dayTime := data.Time.Hour()
	res = append(res, FeatureValue{startIdx + dayTime, 1})

	startIdx += 24

	weekTime := int(data.Time.Weekday())
	res = append(res, FeatureValue{startIdx + weekTime, 1})

	for i, x := range res {
		x.Value = (x.Value - m.Offset) * m.Scale
		res[i] = x
	}

	return res
}

func getUbiquity(envVar string, defaultVal int) int {
	if param := os.Getenv(envVar); param == "" {
		return defaultVal
	} else {
		res, err := strconv.Atoi(param)
		if err != nil {
			fmt.Fprintf(os.Stderr, "invalid %s environment variable", envVar)
			os.Exit(1)
		}
		return res
	}
}
