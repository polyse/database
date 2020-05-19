package collection

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/polyse/database/pkg/filters"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/xujiajun/nutsdb"
)

var (
	dataPrefix       = "d-"
	sourceBucket     = "sources"
	collectionBucket = "collections"
)

// Processor  an interface designed to process and filter incoming data for subsequent
// storing them in a given database collection.
type Processor interface {
	ProcessAndInsertString(data []RawData) error
	ProcessAndGet(query string, limit, offset int) ([]ResponseData, error)
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
	File         string
	MergeTimeout time.Duration
}

// Source structure for domain\article\site\source description
type Source struct {
	Date  time.Time `json:"date" validate:"required"`
	Title string    `json:"title" validate:"required"`
}

type ResponseData struct {
	Source
	Url string `json:"url"`
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

// Name is type to describe collection name in database
type Name string

// NewProcessor function-constructor to SimpleProcessor
func NewSimpleProcessor(
	db *nutsdb.DB,
	colName Name,
	tokenizer filters.Tokenizer,
	textFilters ...filters.Filter,
) (*SimpleProcessor, error) {
	if err := db.Update(func(tx *nutsdb.Tx) error {
		if err := tx.Put(collectionBucket, []byte(colName), []byte(time.Now().String()), 0); err != nil {
			return err
		}
		return nil
	}); err != nil {
		return nil, err
	}
	return &SimpleProcessor{
		db:         db,
		filters:    textFilters,
		tokenizer:  tokenizer,
		colName:    string(colName),
		bucketName: dataPrefix + string(colName),
	}, nil
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
		errChan <- fmt.Errorf("can not save source %s, error %s", data.Url, err)
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

// ProcessAndGet processes the incoming request, dividing it into tokens and filtering,
// after which it finds documents in the specified collection with the maximum number of words from the search query.
// Supports pagination.
func (p *SimpleProcessor) ProcessAndGet(query string, limit, offset int) ([]ResponseData, error) {
	if limit < 1 {
		limit = 10
	}
	if offset < 0 {
		offset = 0
	}
	clearText := p.tokenizer(query, p.filters...)
	return p.findByWords(clearText, limit, offset)
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
	var b bytes.Buffer
	enc := gob.NewEncoder(&b)

	err := enc.Encode(src)
	if err != nil {
		return err
	}
	return p.db.Update(func(tx *nutsdb.Tx) error {
		return tx.Put(sourceBucket, []byte(key), b.Bytes(), 0)
	})
}

func (p *SimpleProcessor) findByWords(keys []string, limit, offset int) (res []ResponseData, err error) {
	log.Debug().
		Strs("search words", keys).
		Int("limit", limit).
		Int("offset", offset).
		Msg("start searching")
	if err = p.db.View(func(tx *nutsdb.Tx) error {
		src, err := findKeys(tx, p.bucketName, keys)
		if err != nil {
			return err
		}
		src = maxKeys(src)
		log.Debug().
			Strs("search words", keys).
			Interface("sources", src).
			Msg("start collect source information")
		res, err = findSources(tx, src)
		if err != nil {
			return err
		}
		return nil
	}); err != nil {
		return nil, err
	}

	log.Debug().
		Strs("search words", keys).
		Int("raw length", len(res)).
		Interface("result", res).
		Msg("data found")

	sort.Slice(res, func(i, j int) bool {
		return res[i].Date.After(res[j].Date)
	})
	if offset >= len(res) {
		offset = 0
	}
	if limit+offset > len(res) {
		limit = len(res) - offset
	}
	res = res[offset : limit+offset]
	log.Debug().
		Strs("search words", keys).
		Int("limit", limit).
		Int("offset", offset).
		Interface("pagination result", res).
		Msg("data found")
	return res, nil
}

func findKeys(tx *nutsdb.Tx, bucketName string, keys []string) (map[string][]string, error) {
	keys = clearDoubleKeys(keys)
	src := make(map[string][]string)
	for i := range keys {
		d, err := tx.SMembers(bucketName, []byte(keys[i]))
		if err != nil {
			if err.Error() == "set not exists" ||
				strings.HasPrefix(err.Error(), "not found bucket:"+bucketName+",key:") {
				log.Warn().Err(err).Str("bucket", bucketName).Str("key", keys[i]).Msg("key not found")
				err = nutsdb.ErrNotFoundKey
			}
			return nil, err
		}
		if err = prepareSet(src, d, keys[i]); err != nil {
			return nil, err
		}
	}
	return src, nil
}

func clearDoubleKeys(keys []string) []string {
	clearMap := make(map[string]struct{})
	var result []string
	for i := range keys {
		if _, ok := clearMap[keys[i]]; !ok {
			clearMap[keys[i]] = struct{}{}
			result = append(result, keys[i])
		}
	}
	return result
}

func prepareSet(src map[string][]string, data [][]byte, word string) error {
	for j := range data {
		var s WordInfo
		r := bytes.NewReader(data[j])
		dec := gob.NewDecoder(r)
		if err := dec.Decode(&s); err != nil {
			return err
		}
		if src[s.Url] == nil {
			src[s.Url] = []string{word}
		} else {
			src[s.Url] = append(src[s.Url], word)
		}
	}
	return nil
}

func maxKeys(input map[string][]string) map[string][]string {
	output := make(map[string][]string)
	max := 0
	for k := range input {
		if len(input[k]) == max {
			output[k] = input[k]
		}
		if len(input[k]) > max {
			max = len(input[k])
			output = make(map[string][]string)
			output[k] = input[k]
		}
	}
	return output
}

func findSources(tx *nutsdb.Tx, src map[string][]string) (res []ResponseData, err error) {
	res = make([]ResponseData, 0, len(src))
	for i := range src {
		e, err := tx.Get(sourceBucket, []byte(i))
		if err != nil {
			return nil, err
		}
		var s Source
		r := bytes.NewReader(e.Value)
		dec := gob.NewDecoder(r)
		if err = dec.Decode(&s); err != nil {
			return nil, err
		}
		res = append(res, ResponseData{
			Source: s,
			Url:    i,
		})
	}
	return res, nil
}

func (p *SimpleProcessor) saveData(ent map[string][]*WordInfo) error {

	p.l.Debug().Interface("data", ent).Msg("start inserting data")

	return p.db.Update(func(tx *nutsdb.Tx) error {
		for i := range ent {
			vals := ent[i]
			data := make([][]byte, 0, len(vals))
			for j := range vals {
				var b bytes.Buffer
				enc := gob.NewEncoder(&b)
				if err := enc.Encode(vals[j]); err != nil {
					return err
				} else {
					data = append(data, b.Bytes())
				}
			}
			if err := tx.SAdd(p.bucketName, []byte(i), data...); err != nil {
				p.l.Err(err).
					Str("key", i).
					Msg("can not SADD to database")
				return err
			}
		}
		return nil
	})
}
