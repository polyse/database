package filter

import "strings"

// Filters is struct for input required filters
type Filters struct {
	Stemming  bool
	ToLower bool
	StopWords bool
}

// TextHandle is returning a slice of tokens after filtering
func TextHandle(text string, f Filters) []string {
	var tokens, output []string

	return output
}
