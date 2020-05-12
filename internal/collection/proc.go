package collection

import (
	"encoding/json"
	"time"

	"github.com/pkg/errors"
	"github.com/polyse/database/pkg/filters"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/xujiajun/nutsdb"
)

var (
	dataPrefix   = "d-"
	sourceBucket = "sources"
)

// Processor  an interface designed to process and filter incoming data for subsequent
// storing them in a given database collection.
type Processor interface {
	ProcessAndInsertString(data []RawData) error
	GetCollectionName() string
}

// SimpleProcessor simple implementation of the Processor interface.
type SimpleProcessor struct {
	tokenizer  filters.Tokenizer
	filters    []filters.Filter
	colName    string
	bucketName string
	db         *nutsdb.DB
	l          zerolog.Logger
}

type wordInfo struct {
	Url string
	Pos []int
}

// Config describes the basic database configuration.
type Config struct {
	File string
}

type Source struct {
	Date  time.Time `json:"date"`
	Title string    `json:"title"`
}

type RawData struct {
	Source
	Url  string `json:"url"`
	Data string `json:"data"`
}

func (i *wordInfo) toJson() []byte {
	b, err := json.Marshal(i)
	if err != nil {
		log.Err(err).Interface("index", wordInfo{}).Msg("can not marshall data")
		return make([]byte, 0)
	}
	return b
}

type Name string

// NewProcessor function-constructor to SimpleProcessor
func NewSimpleProcessor(
	db *nutsdb.DB,
	colName Name,
	tokenizer filters.Tokenizer,
	textFilters ...filters.Filter,
) *SimpleProcessor {
	return &SimpleProcessor{
		db:         db,
		filters:    textFilters,
		tokenizer:  tokenizer,
		colName:    string(colName),
		bucketName: dataPrefix + string(colName),
	}
}

// ProcessAndInsertString changes the input data using the filters specified in this processor,
// and also saves them in a given collection of data bases.
//
// Input format:
// 		[
//          {
//				"url"   : "source1",
//				"date"  : "12.05.2020"
//				"title" : "test title"
//				"data"  : "data1 data2 data3"
//          },
//			{
//				"url"   : "source2",
//				"date"  : "11.06.2020"
//				"title" : "test second title"
//				"data"  : "data2 data3"
//          }
// 		]
// Format after processing:
// 		{
//			"data1" : ["{"url" : "source1", "pos" : [0]}"]
//			"data2" : ["{"url" : "source1", "pos" : [1, 2]}", "{"url" : "source2", "pos" : [0]}"]
//			"data3" : ["{"url" : "source2", "pos" : [1]}"]
//		}
func (p *SimpleProcessor) ProcessAndInsertString(data []RawData) error {
	log.Debug().
		Str("collection in processor", p.GetCollectionName()).
		Msg("processing data")
	parsed := make(map[string][]*wordInfo)
	for k := range data {
		if err := p.saveSource(data[k].Url, Source{Date: data[k].Date, Title: data[k].Title}); err != nil {
			return errors.Wrapf(err, "can not save source %s", data[k].Url)
		}
		clearText := p.tokenizer(data[k].Data, p.filters...)
		sourceMap := buildIndexForOneSource(data[k].Url, clearText)
		for i := range sourceMap {
			if parsed[i] == nil {
				parsed[i] = []*wordInfo{sourceMap[i]}
			} else {
				parsed[i] = append(parsed[i], sourceMap[i])
			}
		}
	}
	return p.saveData(parsed)
}

// GetCollectionName returns the name of the collection specified for this processor.
func (p *SimpleProcessor) GetCollectionName() string {
	return p.colName
}

func buildIndexForOneSource(src string, words []string) map[string]*wordInfo {
	sourceMap := make(map[string]*wordInfo)
	for i := range words {
		if sourceMap[words[i]] == nil {
			sourceMap[words[i]] = &wordInfo{Url: src, Pos: []int{i}}
		} else {
			sourceMap[words[i]].Pos = append(sourceMap[words[i]].Pos, i)
		}
	}
	return sourceMap
}

func (p *SimpleProcessor) saveSource(key string, src Source) error {
	d, err := json.Marshal(src)
	if err != nil {
		return err
	}
	return p.db.Update(func(tx *nutsdb.Tx) error {
		return tx.Put(sourceBucket, []byte(key), d, 0)
	})
}

func (p *SimpleProcessor) saveData(ent map[string][]*wordInfo) error {

	p.l.Debug().Interface("data", ent).Msg("start inserting data")

	return p.db.Update(func(tx *nutsdb.Tx) error {
		for i := range ent {
			vals := ent[i]
			data := make([][]byte, 0, len(vals))
			for j := range vals {
				data = append(data, vals[j].toJson())
			}
			if err := tx.SAdd(p.bucketName, []byte(i), data...); err != nil {
				p.l.Err(err).
					Str("key", i).
					Msg("can not SADD to database")
				if errC := tx.Rollback(); errC != nil {
					p.l.Err(err).Msg("can not rollback transaction")
				}
				return err
			}
		}
		return nil
	})

}
