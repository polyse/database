package proc

import "github.com/polyse/database/internal/db"

type Processor interface {
	ProcessAndInsertString(resource string, data string) error
	GetCollectionName() string
}

type SimpleProcessor struct {
	filters []func()
	repo    db.Repository
}

func NewProcessor(repo db.Repository) *SimpleProcessor {
	return &SimpleProcessor{repo: repo}
}

func (p *SimpleProcessor) ProcessAndInsertString(resource string, data string) error {
	return p.repo.Save(map[string][]string{
		resource: {data},
	})
}

func (p *SimpleProcessor) GetCollectionName() string {
	return p.repo.GetCollectionName()
}
