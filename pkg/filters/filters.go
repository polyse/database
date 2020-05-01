package filters

import (
	"strings"
	"unicode"

	"github.com/kljensen/snowball/english"
	"github.com/zoomio/stopwords"
)

type Filter func(tokens []string) []string

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

func StopWords(tokens []string) []string {
	var output []string
	for _, token := range tokens {
		if !stopwords.IsStopWord(strings.ToLower(token)) {
			output = append(output, token)
		}
	}
	return output
}

func Stemming(tokens []string) []string {
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

func ToLower(tokens []string) []string {
	var output []string
	for _, token := range tokens {
		output = append(output, strings.ToLower(token))
	}
	return output
}
