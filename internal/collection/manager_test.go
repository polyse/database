package collection

import (
	"testing"

	"github.com/polyse/database/pkg/filters"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
)

func TestSimpleProcessor_GetCollectionName(t *testing.T) {
	type fields struct {
		filters   []filters.Filter
		colName   string
		tokenizer filters.Tokenizer
	}
	tests := []struct {
		name   string
		fields fields
		want   string
	}{
		{
			name: "Normal Test",
			fields: fields{
				colName: nutColl,
			},
			want: nutColl,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &SimpleProcessor{
				filters:   tt.fields.filters,
				colName:   tt.fields.colName,
				tokenizer: tt.fields.tokenizer,
			}
			if got := p.GetCollectionName(); got != tt.want {
				t.Errorf("GetCollectionName() = %v, want %v", got, tt.want)
			}
		})
	}
}

type processorManagerTestSuite struct {
	suite.Suite
	prm *Manager
	tr  *MockProcessor
	tr2 *MockProcessor
}

func TestStartProcessorManagerSuit(t *testing.T) {
	suite.Run(t, new(processorManagerTestSuite))
}

func (pts *processorManagerTestSuite) SetupTest() {
	testProc := new(MockProcessor)

	testProc.
		On("ProcessAndInsertString", mock.Anything, mock.Anything).
		Return(nil).
		On("GetCollectionName").
		Return("testCollection")

	testProc2 := new(MockProcessor)
	testProc2.
		On("ProcessAndInsertString", mock.Anything, mock.Anything).
		Return(nil).
		On("GetCollectionName").
		Return("secondTestCollection")

	pts.prm = NewManagerWithProc(testProc)

	pts.prm.AddProcessor(testProc2)
	pts.tr = testProc
	pts.tr2 = testProc2
}

func (pts *processorManagerTestSuite) TestSimpleProcessorManager_AddProcessors() {
	pts.Len(pts.prm.processors, 2)
	pts.prm.AddProcessor(
		NewSimpleProcessor(
			nil,
			"testCollection3",
			filters.FilterText,
			filters.StemmAndToLower,
			filters.StopWords,
		),
	)
	pts.Len(pts.prm.processors, 3)
}

func (pts *processorManagerTestSuite) TestSimpleProcessorManager_GetProcessor() {
	p, err := pts.prm.GetProcessor(
		"testCollection",
	)
	pts.NoError(err)
	pts.NoError(p.ProcessAndInsertString([]RawData{{Url: "test", Data: "data"}}))
	pts.tr.AssertCalled(pts.T(), "ProcessAndInsertString", []RawData{{Url: "test", Data: "data"}})
	pts.tr2.AssertNotCalled(pts.T(), "ProcessAndInsertString", mock.Anything, mock.Anything)
}

type MockProcessor struct {
	mock.Mock
}

// GetCollectionName provides a mock function with given fields:
func (_m *MockProcessor) GetCollectionName() string {
	ret := _m.Called()

	var r0 string
	if rf, ok := ret.Get(0).(func() string); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(string)
	}

	return r0
}

// ProcessAndInsertString provides a mock function with given fields: data
func (_m *MockProcessor) ProcessAndInsertString(data []RawData) error {
	ret := _m.Called(data)

	var r0 error
	if rf, ok := ret.Get(0).(func([]RawData) error); ok {
		r0 = rf(data)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}
