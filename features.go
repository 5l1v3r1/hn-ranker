package main

import (
	"strings"
	"time"
	"unicode"
)

// TimeOfDay is an integer between 0 and 23
// indicating the hour of a given timestamp.
type TimeOfDay int

// TimeOfWeek is an integer between 0 and 6
// indicating the day of the week of a given
// timestamp, starting on Sunday.
type TimeOfWeek int

// Features describes the format of a
// feature vector.
type Features struct {
	TitleKeywords   []string
	ContentKeywords []string
	HostNames       []string
}

// A Sample contains the raw information
// about a story.
type Sample struct {
	Title    string
	Content  string
	HostName string

	DayTime  TimeOfDay
	WeekTime TimeOfWeek
}

func NewSample(title, content, hostName string, time time.Time) *Sample {
	return &Sample{
		Title:    title,
		Content:  content,
		HostName: hostName,
		DayTime:  TimeOfDay(time.Hour()),
		WeekTime: TimeOfWeek(time.Weekday()),
	}
}

func (s *Sample) FeatureVector(features *Features) []float64 {
	res := make([]float64, 0, len(features.TitleKeywords)+len(features.ContentKeywords)+
		len(features.HostNames)+24+7)
	titleKeywords := extractKeywords(s.Title)
	contentKeywords := extractKeywords(s.Content)
	for i, x := range features.ContentKeywords {
		res[i] = contentKeywords[x]
	}
	for i, x := range features.TitleKeywords {
		res[i+len(features.ContentKeywords)] = titleKeywords[x]
	}
	for i, host := range features.HostNames {
		if host == s.HostName {
			res[i+len(features.ContentKeywords)+len(features.TitleKeywords)] = 1
		}
	}
	offset := len(features.TitleKeywords) + len(features.ContentKeywords) +
		len(features.HostNames)
	res[offset+int(s.DayTime)] = 1
	res[offset+24+int(s.WeekTime)] = 1
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
