// Package proc is intended for separation, processing of incoming data using the specified filters
// for each collection for subsequent data saving to the database.
package collection

import (
	"fmt"
	"sync"

	"github.com/rs/zerolog/log"
)

// SimpleProcessorManager structure a simple implementation of the ProcessManager interface
// is a processor map, where the key is the name of the collection and the value is the processor itself.
type SimpleProcessorManager struct {
	sync.RWMutex
	processors map[string]Processor
}

// NewSimpleProcessorManager function-constructor of SimpleProcessorManager
func NewSimpleProcessorManager() *SimpleProcessorManager {
	return &SimpleProcessorManager{processors: make(map[string]Processor)}
}

// NewSimpleProcessorManagerWithProc function-constructor of SimpleProcessorManager with a given Processor.
func NewSimpleProcessorManagerWithProc(proc Processor) *SimpleProcessorManager {
	spm := NewSimpleProcessorManager()
	spm.AddProcessor(proc)
	return spm
}

// AddProcessor adding more processor to manager.
func (spm *SimpleProcessorManager) AddProcessor(proc ...Processor) {
	for i := range proc {
		log.Debug().Str("processor collection", proc[i].GetCollectionName()).Msg("adding processors")
	}

	spm.Lock()
	defer spm.Unlock()
	for i := range proc {
		spm.processors[proc[i].GetCollectionName()] = proc[i]
	}
}

// ProcessAndInsertString selects the necessary processor for this collection
// and transfers data to it for subsequent processing and storage.
func (spm *SimpleProcessorManager) ProcessAndInsertString(data []RawData, colName string) error {
	log.Debug().Str("collection name", colName).Msg("manager, start inserting data")
	spm.RLock()
	val, ok := spm.processors[colName]
	spm.RUnlock()
	if ok {
		return val.ProcessAndInsertString(data)
	}
	return fmt.Errorf("collection named %s does not exist", colName)
}
