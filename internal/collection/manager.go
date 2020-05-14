// Package proc is intended for separation, processing of incoming data using the specified filters
// for each collection for subsequent data saving to the database.
package collection

import (
	"errors"
	"sync"

	"github.com/rs/zerolog/log"
)

var (
	ErrCollectionNotExist = errors.New("collection does not exist")
)

// Manager structure a simple implementation of the ProcessManager interface
// is a processor map, where the key is the name of the collection and the value is the processor itself.
type Manager struct {
	sync.RWMutex
	processors map[string]Processor
}

// NewManager function-constructor of Manager
func NewManager() *Manager {
	return &Manager{processors: make(map[string]Processor)}
}

// NewManagerWithProc function-constructor of Manager with a given Processor.
func NewManagerWithProc(proc Processor) *Manager {
	spm := NewManager()
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

func (spm *SimpleProcessorManager) ProcessAndGetData(colName, query string, limit, offset int) ([]ResponseData, error) {
	log.Debug().Str("collection name", colName).Msg("manager, start finding data")
	spm.RLock()
	val, ok := spm.processors[colName]
	spm.RUnlock()
	if ok {
		return val.ProcessAndGet(query, limit, offset)
	}
	return nil, fmt.Errorf("collection named %s does not exist", colName)
}
