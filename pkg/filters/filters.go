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
	tokens := strings.Fields(text)

	for i, token := range tokens {
		tokens[i] = strings.TrimFunc(token, func(r rune) bool {
			return !unicode.IsLetter(r) && !unicode.IsNumber(r)
		})
	}

	for _, filter := range filters {
		tokens = filter(tokens)
	}

	return tokens
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

// Stemm stemm tokens
func Stemm(tokens []string) []string {
	var output []string
	for _, token := range tokens {
		stemmedToken := english.Stem(token, false)
		if len(stemmedToken) != len(token) {
			token = token[strings.Index(strings.ToLower(token), stemmedToken):len(stemmedToken)]
		}
		output = append(output, token)
	}
	return output
}

// ToLower apply alower case for tokens
func ToLower(tokens []string) []string {
	var output []string
	for _, token := range tokens {
		output = append(output, strings.ToLower(token))
	}
	return output
}
