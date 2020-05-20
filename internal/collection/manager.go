// Package proc is intended for separation, processing of incoming data using the specified filters
// for each collection for subsequent data saving to the database.
package collection

import (
	"encoding/json"
	"errors"
	"github.com/polyse/database/pkg/filters"
	"github.com/rs/zerolog/log"
	"github.com/xujiajun/nutsdb"
	"sync"
)

var (
	ErrCollectionNotExist = errors.New("collection does not exist")
)

// Manager structure a simple implementation of the ProcessManager interface
// is a processor map, where the key is the name of the collection and the value is the processor itself.
type Manager struct {
	sync.RWMutex
	processors map[string]Processor
	db         *nutsdb.DB
}

// NewManager function-constructor of Manager
func NewManager(db *nutsdb.DB) *Manager {
	m, err := loadCollections(db)
	if err != nil {
		log.Err(err).Msg("can not load collections")
		m = make(map[string]Processor)
	}
	return &Manager{processors: m, db: db}
}

// NewManagerWithProc function-constructor of Manager with a given Processor.
func NewManagerWithProc(db *nutsdb.DB, proc Processor) *Manager {
	spm := NewManager(db)
	spm.AddProcessor(proc)
	return spm
}

// AddProcessor adding more processor to manager.
func (spm *Manager) AddProcessor(proc ...Processor) {
	for i := range proc {
		log.Debug().Str("processor collection", proc[i].GetCollectionName()).Msg("adding processors")
	}

	spm.Lock()
	defer spm.Unlock()
	for i := range proc {
		spm.processors[proc[i].GetCollectionName()] = proc[i]
	}
}

// GetProcessor returns processor to the specific collection.
func (spm *Manager) GetProcessor(colName string) (Processor, error) {
	log.Debug().Str("collection name", colName).Msg("manager, start inserting data")
	spm.RLock()
	val, ok := spm.processors[colName]
	spm.RUnlock()
	if ok {
		return val, nil
	}
	return nil, ErrCollectionNotExist
}

func (spm *Manager) InitNewProc(colName string, tokenizer string, filter ...string) (Processor, error) {
	return newSimpleProcessorByStrings(spm.db, Name(colName), tokenizer, filter...)
}

func (spm *Manager) GetAllCollectionsInfo() (map[string]Metadata, error) {
	log.Debug().Msg("getting all collections")
	collections := make(map[string][]byte)
	if err := spm.db.View(func(tx *nutsdb.Tx) error {
		e, err := tx.GetAll(collectionBucket)
		if err != nil {
			return err
		}
		for i := range e {
			collections[string(e[i].Key)] = e[i].Value
		}
		return nil
	}); err != nil {
		return nil, err
	}
	log.Debug().Msg("process get all collections data")
	metaMap := make(map[string]Metadata)
	for i := range collections {
		var p Metadata
		if err := json.Unmarshal(collections[i], &p); err != nil {
			return nil, err
		}
		metaMap[i] = p
	}
	log.Debug().Interface("all collections", metaMap).Msg("getting collections done")
	return metaMap, nil
}

func loadCollections(db *nutsdb.DB) (map[string]Processor, error) {
	log.Debug().Msg("start loading collections")
	collections := make(map[string][]byte)
	if err := db.View(func(tx *nutsdb.Tx) error {
		e, err := tx.GetAll(collectionBucket)
		if err != nil {
			return err
		}
		for i := range e {
			collections[string(e[i].Key)] = e[i].Value
		}
		return nil
	}); err != nil {
		return nil, err
	}
	procMap := make(map[string]Processor)
	for i := range collections {
		var p Metadata
		if err := json.Unmarshal(collections[i], &p); err != nil {
			return nil, err
		}
		log.Debug().Str("collection name", i).Interface("metadata", p).Msg("collection loaded")
		var f []filters.Filter
		for _, t := range p.ColFilters {
			f = append(f, filterMap[t])
		}
		sp, err := NewSimpleProcessor(db, Name(i), tokenizerMap[p.Tokenizer], f...)
		if err != nil {
			return nil, err
		}
		procMap[i] = sp
	}
	return procMap, nil
}

var filterMap = map[string]filters.Filter{
	"github.com/polyse/database/pkg/filters.StemmAndToLower": filters.StemmAndToLower,
	"github.com/polyse/database/pkg/filters.StopWords":       filters.StopWords,
}
var tokenizerMap = map[string]filters.Tokenizer{
	"github.com/polyse/database/pkg/filters.FilterText": filters.FilterText,
}
