package collection

import (
	"github.com/polyse/database/pkg/filters"
	"testing"

	"github.com/polyse/database/test/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
)

func TestSimpleProcessor_GetCollectionName(t *testing.T) {
	type fields struct {
		filters []filters.Filter
		repo    Repository
	}
	tests := []struct {
		name   string
		fields fields
		want   string
	}{
		{
			name: "Normal Test",
			fields: fields{
				repo: NewNutRepo("testCollection", nil),
			},
			want: "testCollection",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &SimpleProcessor{
				filters: tt.fields.filters,
				repo:    tt.fields.repo,
			}
			if got := p.GetCollectionName(); got != tt.want {
				t.Errorf("GetCollectionName() = %v, want %v", got, tt.want)
			}
		})
	}
}

type processorTestSuite struct {
	suite.Suite
	pr Processor
	tr *mocks.Repository
}

func TestStartProcessorSuit(t *testing.T) {
	suite.Run(t, new(processorTestSuite))
}

func (pts *processorTestSuite) SetupTest() {
	testRepo := new(mocks.Repository)

	testRepo.
		On("GetCollectionName").
		Return("testCollection").
		On("Save", mock.Anything).
		Return(nil)

	pts.pr = NewSimpleProcessor(testRepo, filters.FilterText, filters.StemmAndToLower, filters.StopWords)
	pts.tr = testRepo
}

func (pts *processorTestSuite) TestSimpleProcessor_ProcessAndInsertString() {
	assert.NoError(pts.T(), pts.pr.ProcessAndInsertString(map[string]string{"test": "is Data map"}))
	pts.tr.AssertCalled(pts.T(), "Save", map[string][]string{"data": {"test"}, "map": {"test"}})
}

type processorManagerTestSuite struct {
	suite.Suite
	prm *SimpleProcessorManager
	tr  *mocks.Processor
	tr2 *mocks.Processor
}

func TestStartProcessorManagerSuit(t *testing.T) {
	suite.Run(t, new(processorManagerTestSuite))
}

func (pts *processorManagerTestSuite) SetupTest() {
	testProc := new(mocks.Processor)

	testProc.
		On("ProcessAndInsertString", mock.Anything, mock.Anything).
		Return(nil).
		On("GetCollectionName").
		Return("testCollection")

	testProc2 := new(mocks.Processor)
	testProc2.
		On("ProcessAndInsertString", mock.Anything, mock.Anything).
		Return(nil).
		On("GetCollectionName").
		Return("secondTestCollection")

	pts.prm = NewSimpleProcessorManagerWithProc(testProc)

	pts.prm.AddProcessor(testProc2)
	pts.tr = testProc
	pts.tr2 = testProc2
}

func (pts *processorManagerTestSuite) TestSimpleProcessorManager_AddProcessors() {
	pts.Len(pts.prm.processors, 2)
	pts.prm.AddProcessor(
		NewSimpleProcessor(
			NewNutRepo("testCollection3", nil),
			filters.FilterText,
			filters.StemmAndToLower,
			filters.StopWords,
		),
	)
	pts.Len(pts.prm.processors, 3)
}

func (pts *processorManagerTestSuite) TestSimpleProcessorManager_ProcessAndInsertString() {
	assert.NoError(pts.T(), pts.prm.ProcessAndInsertString(map[string]string{"test": "data"}, "testCollection"))
	pts.tr.AssertCalled(pts.T(), "ProcessAndInsertString", map[string]string{"test": "data"})
	pts.tr2.AssertNotCalled(pts.T(), "ProcessAndInsertString", mock.Anything, mock.Anything)
}
