package collection

import (
	"encoding/json"
	"sync"
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

// Config describes the basic database configuration.
type Config struct {
	File string
}

// Source structure for domain\article\site\source description
type Source struct {
	Date  time.Time `json:"date" validate:"required"`
	Title string    `json:"title" validate:"required"`
}

// RawData structure for json data description
type RawData struct {
	Source `json:"source" validate:"required,dive"`
	Url    string `json:"url" validate:"required,url"`
	Data   string `json:"data" validate:"required"`
}

// WordInfo structure for describing positions of tokens in the text at a given url
type WordInfo struct {
	Url string
	Pos []int
}

func (i *WordInfo) toJson() ([]byte, error) {
	b, err := json.Marshal(i)
	if err != nil {
		log.Err(err).Interface("index", i).Msg("can not marshall data")
		return nil, err
	}
	return b, err
}

// Name is type to describe collection name in database
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
//    [
//      {
//        "url"   : "source1",
//        "date"  : "12.05.2020"
//        "title" : "test title"
//        "data"  : "data1 data2 data3"
//      },
//      {
//        "url"   : "source2",
//        "date"  : "11.06.2020"
//        "title" : "test second title"
//        "data"  : "data2 data3"
//      }
//    ]
// Format after processing:
//    {
//      "data1" : ["{"url" : "source1", "pos" : [0]}"],
//      "data2" : ["{"url" : "source1", "pos" : [1, 2]}", "{"url" : "source2", "pos" : [0]}"],
//      "data3" : ["{"url" : "source2", "pos" : [1]}"],
//    }
func (p *SimpleProcessor) ProcessAndInsertString(data []RawData) error {
	log.Debug().
		Str("collection in processor", p.GetCollectionName()).
		Msg("processing data")
	parsed := make(map[string][]*WordInfo)
	dataCh := make(chan map[string]*WordInfo, len(data))
	errCh := make(chan error, 1)
	var wg sync.WaitGroup

	for k := range data {
		wg.Add(1)
		go func(wr *sync.WaitGroup, data RawData, errChan chan<- error, dataChan chan<- map[string]*WordInfo) {
			defer wg.Done()
			p.asyncProcessData(data, errCh, dataCh)
		}(&wg, data[k], errCh, dataCh)
	}
	go func(wg *sync.WaitGroup, dataChan chan map[string]*WordInfo, errChan chan error) {
		wg.Wait()
		close(errCh)
		close(dataCh)
	}(&wg, dataCh, errCh)
ReadLoop:
	for {
		select {
		case d, ok := <-dataCh:
			if !ok {
				break ReadLoop
			}
			for i := range d {
				if parsed[i] == nil {
					parsed[i] = []*WordInfo{d[i]}
				} else {
					parsed[i] = append(parsed[i], d[i])
				}
			}
		case err, ok := <-errCh:
			if ok {
				return err
			}
		}
	}

	return p.saveData(parsed)
}

func (p *SimpleProcessor) asyncProcessData(data RawData, errChan chan<- error, dataChan chan<- map[string]*WordInfo) {
	if err := p.saveSource(data.Url, Source{Date: data.Date, Title: data.Title}); err != nil {
		errChan <- errors.Wrapf(err, "can not save source %s", data.Url)
		return
	}
	clearText := p.tokenizer(data.Data, p.filters...)
	sourceMap := buildIndexForOneSource(data.Url, clearText)
	dataChan <- sourceMap
}

// GetCollectionName returns the name of the collection specified for this processor.
func (p *SimpleProcessor) GetCollectionName() string {
	return p.colName
}

func buildIndexForOneSource(src string, words []string) map[string]*WordInfo {
	sourceMap := make(map[string]*WordInfo)
	for i := range words {
		if sourceMap[words[i]] == nil {
			sourceMap[words[i]] = &WordInfo{Url: src, Pos: []int{i}}
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

func (p *SimpleProcessor) saveData(ent map[string][]*WordInfo) error {

	p.l.Debug().Interface("data", ent).Msg("start inserting data")

	return p.db.Update(func(tx *nutsdb.Tx) error {
		for i := range ent {
			vals := ent[i]
			data := make([][]byte, 0, len(vals))
			for j := range vals {
				if b, err := vals[j].toJson(); err != nil {
					return p.rollbackTransaction(tx, err)
				} else {
					data = append(data, b)
				}
			}
			if err := tx.SAdd(p.bucketName, []byte(i), data...); err != nil {
				p.l.Err(err).
					Str("key", i).
					Msg("can not SADD to database")
				return p.rollbackTransaction(tx, err)
			}
		}
		return nil
	})
}

func (p *SimpleProcessor) rollbackTransaction(tx *nutsdb.Tx, err error) error {
	if errC := tx.Rollback(); errC != nil {
		p.l.Err(err).Msg("can not rollback transaction")
	}
	return err
}
