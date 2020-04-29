package filter

import (
	"strings"

	"github.com/kljensen/snowball/english"
	"github.com/zoomio/stopwords"
)

// Filters is struct for input required filters
type Filters struct {
	Stemming  bool
	ToLower   bool
	StopWords bool
}

// TextHandle is returning a slice of tokens after filtering
func TextHandle(text string, f Filters) []string {
	var output []string

	if f.ToLower {
		text = strings.ToLower(text)
	}

	output = strings.Fields(text)

	if f.StopWords {
		output = output.stopWordsDelete()
	}

	if f.Stemming {
		output = output.stemming()
	}

	return output
}

func (tokens []string) stopWordsDelete() []string {
	var output []string
	for _, token := range tokens {
		if !stopwords.IsStopWord(token) {
			output = append(output, token)
		}
	}
	return output
}

func (tokens []string) stemming() []string {
	var output []string
	for _, token := range tokens {
		output = append(output, english.Stem(token, false))
	}
	return output
}
