package main

import (
	"fmt"
	"strings"
	"unicode"
)

func wordFrequencyCount(s string) map[string]int {

	wordCount := make(map[string]int)

	// a custom function to check for delimeters(white space and any punctuation)
	f := func(r rune) bool {
		return unicode.IsSpace(r) || unicode.IsPunct(r)
	}

	for _, word := range strings.FieldsFunc(s, f) {
		lowerWord := strings.ToLower(word)
		wordCount[lowerWord]++
	}

	return wordCount
}

func main() {
	text := "some text and punctuations!"
	output := wordFrequencyCount(text)
	fmt.Println(output)
}
