package filter

import (
	"strings"

	"github.com/reiver/go-porterstemmer"
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
		output = stopWordsDelete(output)
	}

	if f.Stemming {
		output = stemming(output)
	}

	return output
}

func stopWordsDelete(tokens []string) []string {
	var output []string
	for _, token := range tokens {
		if !stopwords.IsStopWord(token) {
			output = append(output, token)
		}
	}
	return output
}

func stemming(tokens []string) []string {
	var output []string
	for _, token := range tokens {
		output = append(output, string(porterstemmer.StemWithoutLowerCasing([]rune(token))))
	}
	return output
}
