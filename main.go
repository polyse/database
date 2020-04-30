package main

import (
	"fmt"
	"strings"

	"github.com/polyse/database/cmd/filters/stemming"
	"github.com/polyse/database/cmd/filters/stopwords"
	"github.com/polyse/database/cmd/filters/tolower"
)

func main() {
	output := strings.Fields("the cup OF BLacking tea with sugar ")
	output = tolower.Handle(output)
	output = stopwords.Handle(output)
	output = stemming.Handle(output)
	fmt.Println(output)
}
