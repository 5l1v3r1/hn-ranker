package main

import (
	"strings"
	"time"
	"unicode"
)

// A FeatureMap describes how to map data from
// a story to a vector of scalers.
type FeatureMap struct {
	TitleKeywords   []string
	ContentKeywords []string
	HostNames       []string
}

func (f *FeatureMap) VectorSize() int {
	// Room for all tokens, the day of week, and the hour of day.
	return len(f.TitleKeywords) + len(f.ContentKeywords) + len(f.HostNames) + 24 + 7
}

// StoryData contains the raw data of a story,
// before it is converter into a feature vector.
type StoryData struct {
	Title    string
	Content  string
	HostName string
	Time     time.Time
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

	return res
}

func extractKeywords(content string) map[string]float64 {
	counts := map[string]int{}
	extractLetterKeywords(content, counts)
	for _, word := range strings.Fields(content) {
		if !isOnlyLetters(word) {
			counts[word]++
		}
	}

	var totalCount int
	for _, count := range counts {
		totalCount += count
	}

	res := map[string]float64{}
	for word, count := range counts {
		res[word] = float64(count) / float64(totalCount)
	}

	return res
}

func extractLetterKeywords(content string, res map[string]int) {
	currentWord := ""
	for _, chr := range content {
		if unicode.IsLetter(chr) {
			currentWord += string(chr)
		} else {
			if currentWord != "" {
				res[strings.ToLower(currentWord)]++
				currentWord = ""
			}
		}
	}
	if currentWord != "" {
		res[strings.ToLower(currentWord)]++
	}
}

func isOnlyLetters(s string) bool {
	for _, r := range s {
		if !unicode.IsLetter(r) {
			return false
		}
	}
	return true
}
