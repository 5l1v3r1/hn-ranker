package main

import (
	"strings"
	"unicode"
)

// Features describes the format of a
// feature vector.
type Features struct {
	TitleKeywords   []string
	ContentKeywords []string
	HostNames       []string
}

func ExtractKeywords(content string) map[string]float64 {
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
