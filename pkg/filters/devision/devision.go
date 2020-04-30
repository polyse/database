package devision

import (
	"strings"
	"unicode"
)

func Handle(text string) []string {
	var output []string
	output = strings.Fields(text)
	for i, token := range output {
		output[i] = strings.TrimFunc(token, func(r rune) bool {
			return !unicode.IsLetter(r) && !unicode.IsNumber(r)
		})
	}
	return output
}
