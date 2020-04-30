package stemming

import (
	"github.com/reiver/go-porterstemmer"
)

func Handle(tokens []string) []string {
	var output []string
	for _, token := range tokens {
		output = append(output, string(porterstemmer.StemWithoutLowerCasing([]rune(token))))
	}
	return output
}
