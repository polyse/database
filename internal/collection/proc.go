package collection

import (
	"encoding/json"

	"github.com/polyse/database/pkg/filters"
	"github.com/rs/zerolog/log"
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
	repo      Repository
	colName   string
}

type Index struct {
	source string
	pos    []int
}

func (i *Index) GetBytes() []byte {
	b, err := json.Marshal(i)
	if err != nil {
		log.Err(err).Interface("index", Index{}).Msg("can not marshall data")
		return make([]byte, 0)
	}
	return b
}

type Name string

// NewProcessor function-constructor to SimpleProcessor
func NewSimpleProcessor(repo Repository, colName Name, tokenizer filters.Tokenizer, textFilters ...filters.Filter) *SimpleProcessor {
	return &SimpleProcessor{repo: repo, filters: textFilters, tokenizer: tokenizer, colName: string(colName)}
}

// ProcessAndInsertString changes the input data using the filters specified in this processor,
// and also saves them in a given collection of data bases.
//
// Input format:
// 		{
//			"source1" : "data1 data2 data2"
//			"source2" : "data2 data3"
// 		}
// Format after processing:
// 		{
//			"data1" : ["{"source1" : [0]}"]
//			"data2" : ["{"source1" : [1, 2]}", "{"source2" : [0]}"]
//			"data3" : ["{"source2" : [1]}"]
//		}
func (p *SimpleProcessor) ProcessAndInsertString(data map[string]string) error {
	log.Debug().
		Str("collection in processor", p.GetCollectionName()).
		Msg("processing data")
	parsed := make(map[string][]ByteArr)
	for k := range data {
		clearText := p.tokenizer(data[k], p.filters...)
		sourceMap := buildIndexForOneSource(k, clearText)
		for i := range sourceMap {
			if parsed[i] == nil {
				parsed[i] = []ByteArr{sourceMap[i]}
			} else {
				parsed[i] = append(parsed[i], sourceMap[i])
			}
		}
	}
	return p.repo.Save(parsed)
}

func buildIndexForOneSource(fn string, src []string) map[string]*Index {
	sourceMap := make(map[string]*Index)
	for i := range src {
		if sourceMap[src[i]] == nil {
			sourceMap[src[i]] = &Index{fn, []int{i}}
		} else {
			sourceMap[src[i]].pos = append(sourceMap[src[i]].pos, i)
		}
	}
	return sourceMap
}

// GetCollectionName returns the name of the collection specified for this processor.
func (p *SimpleProcessor) GetCollectionName() string {
	return p.colName
}
