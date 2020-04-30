package stopwords

import (
	"strings"

	"github.com/zoomio/stopwords"
)

func Handle(tokens []string) []string {
	var output []string
	for _, token := range tokens {
		if !stopwords.IsStopWord(strings.ToLower(token)) {
			output = append(output, token)
		}
	}
	return output
}
