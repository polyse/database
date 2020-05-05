package filters

import (
	"strings"
	"unicode"

	"github.com/kljensen/snowball/english"
	"github.com/zoomio/stopwords"
)

// Filter is type for input sort functions as parameters to FilterText
type Filter func(tokens []string) []string

// FilterText divide text to tokens, trim tokens and apply filters to tokens
func FilterText(text string, filters ...Filter) []string {
	tokens := strings.FieldsFunc(text, func(c rune) bool {
		return !unicode.IsLetter(c) && !unicode.IsNumber(c) && c != '\'' && c != '-'
	})

	var output []string
	for _, token := range tokens {
		if token != "'" && token != "-" {
			output = append(output, token)
		}
	}

	for _, filter := range filters {
		output = filter(output)
	}

	return output
}

// StopWords remove stop words from tokens
func StopWords(tokens []string) []string {
	var output []string
	for _, token := range tokens {
		if !stopwords.IsStopWord(strings.ToLower(token)) {
			output = append(output, token)
		}
	}
	return output
}

// StemmAndToLower stemm tokens and change them to lower case
func StemmAndToLower(tokens []string) []string {
	output := make([]string, 0, len(tokens))
	for _, token := range tokens {
		token := english.Stem(token, false)
		output = append(output, token)
	}
	return output
}
