package tolower

import "strings"

func Handle(tokens []string) []string {
	var output []string
	for _, token := range tokens {
		output = append(output, strings.ToLower(token))
	}
	return output
}
