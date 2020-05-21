package collection

import (
	"github.com/xujiajun/nutsdb"
	"os"
	"testing"

	"github.com/polyse/database/pkg/filters"

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
	prm    *Manager
	tr     *SimpleProcessor
	tr2    *SimpleProcessor
	nutsDb *nutsdb.DB
}

func TestStartProcessorManagerSuit(t *testing.T) {
	suite.Run(t, new(processorManagerTestSuite))
}

func (pts *processorManagerTestSuite) SetupTest() {
	opt := nutsdb.DefaultOptions
	opt.Dir = dbDir
	nutsDb, err := nutsdb.Open(opt)
	if err != nil {
		panic(err)
	}
	pts.nutsDb = nutsDb
	testProc, err := NewSimpleProcessor(nutsDb, Name("test1"), filters.FilterText, filters.StemmAndToLower)
	if err != nil {
		panic(err)
	}
	testProc2, err := NewSimpleProcessor(nutsDb, Name("test2"), filters.FilterText, filters.StopWords)
	if err != nil {
		panic(err)
	}

	pts.prm = NewManagerWithProc(nutsDb, testProc)

	pts.prm.AddProcessor(testProc2)
	pts.tr = testProc
	pts.tr2 = testProc2
}

func (pts *processorManagerTestSuite) TearDownSuite() {
	_ = pts.nutsDb.Close()
	if err := os.RemoveAll(dbDir); err != nil {
		panic(err)
	}
}

func (pts *processorManagerTestSuite) TestSimpleProcessorManager_LoadCollection() {
	testNormalProc, err := NewSimpleProcessor(pts.nutsDb, "test", filters.FilterText)
	if err != nil {
		panic(err)
	}
	pts.Len(pts.prm.processors, 2)
	pts.prm.AddProcessor(testNormalProc)
	pts.Len(pts.prm.processors, 3)
	if err := pts.nutsDb.Close(); err != nil {
		panic(err)
	}
	opt := nutsdb.DefaultOptions
	opt.Dir = dbDir
	nutsDb, err := nutsdb.Open(opt)
	pts.NoError(err)
	man := NewManager(nutsDb)
	p, err := man.GetProcessor("test")
	pts.NoError(err)
	pts.IsType(testNormalProc, p)
	pts.Equal(testNormalProc.colName, p.GetCollectionName())
	meta, err := man.GetAllCollectionsInfo()
	pts.NoError(err)
	pts.Len(meta, 3)
	_, ok := meta["test1"]
	pts.True(ok)
	_, ok = meta["test2"]
	pts.True(ok)
}

func (pts *processorManagerTestSuite) TestSimpleProcessorManager_GetProcessor() {
	p, err := pts.prm.GetProcessor(
		"test1",
	)
	pts.NoError(err)
	pts.Equal("test1", p.GetCollectionName())
}
