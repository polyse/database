package proc

import (
	"github.com/polyse/database/internal/db"
	"github.com/polyse/database/pkg/filters"
)

// Processor  an interface designed to process and filter incoming data for subsequent
// storing them in a given database collection.
type Processor interface {
	ProcessAndInsertString(data map[string]string) error
	GetCollectionName() string
}

// SimpleProcessor simple implementation of the Processor interface.
type SimpleProcessor struct {
	tokenizer filters.Tokenizer
	filters   []filters.Filter
	repo      db.Repository
}

// NewProcessor function-constructor to SimpleProcessor
func NewProcessor(repo db.Repository, tokenizer filters.Tokenizer, textFilters ...filters.Filter) *SimpleProcessor {
	return &SimpleProcessor{repo: repo, filters: textFilters, tokenizer: tokenizer}
}

// ProcessAndInsertString changes the input data using the filters specified in this processor,
// and also saves them in a given collection of data bases.
//
// Input format:
// 		 {
//			 "source1" : "data1 data2"
//			 "source2" : "data2 data3"
// 		 }
// Format after processing:
// 		{
//			"data1" : ["source1"]
//			"data2" : ["source1", "source2"]
//			"data3" : ["source2"]
//		}
func (p *SimpleProcessor) ProcessAndInsertString(data map[string]string) error {
	parsed := make(map[string][]string)
	for k := range data {
		clearText := filters.FilterText(data[k], p.filters...)
		for i := range clearText {
			if parsed[clearText[i]] == nil {
				parsed[clearText[i]] = []string{k}
			} else {
				parsed[clearText[i]] = append(parsed[clearText[i]], k)
			}
		}
	}
	return p.repo.Save(parsed)
}

// GetCollectionName returns the name of the collection specified for this processor.
func (p *SimpleProcessor) GetCollectionName() string {
	return p.repo.GetCollectionName()
}
