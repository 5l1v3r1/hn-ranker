package hnclass

import (
	"strings"
	"unicode"
)

func extractKeywords(content string) map[string]float64 {
	counts := map[string]int{}
	extractLetterKeywords(content, counts)

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
	wordStart := 0
	runes := []rune(content)
	for i, r := range runes {
		if !unicode.IsLetter(r) {
			if i > wordStart {
				word := string(runes[wordStart:i])
				res[strings.ToLower(word)]++
				wordStart = i + 1
			} else {
				wordStart = i + 1
			}
		}
	}
	if wordStart < len(runes) {
		word := string(runes[wordStart:])
		res[strings.ToLower(word)]++
	}
}
