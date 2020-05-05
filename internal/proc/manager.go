package proc

import "fmt"

type ProcessorManager interface {
	AddProcessor(proc ...Processor)
	ProcessAndInsertString(colName string, resource string, toProcess string) error
}

type SimpleProcessorManager map[string]Processor

func NewSimpleProcessorManager() SimpleProcessorManager {
	return make(map[string]Processor)
}

func NewSimpleProcessorManagerWithProc(proc Processor) SimpleProcessorManager {
	spm := NewSimpleProcessorManager()
	spm.AddProcessor(proc)
	return spm
}

func (spm SimpleProcessorManager) AddProcessor(proc ...Processor) {
	for i := range proc {
		spm[proc[i].GetCollectionName()] = proc[i]
	}
}

func (spm SimpleProcessorManager) ProcessAndInsertString(colName, resource, toProcess string) error {
	if val, ok := spm[colName]; ok {
		return val.ProcessAndInsertString(resource, toProcess)
	}
	return fmt.Errorf("collection named %s does not exist", colName)
}
