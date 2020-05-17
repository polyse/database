package collection

import (
	"encoding/json"
	"os"
	"testing"
	"time"

	"github.com/polyse/database/pkg/filters"

	"github.com/stretchr/testify/suite"
	"github.com/xujiajun/nutsdb"
)

var (
	dbDir   = "nutsdb-test"
	nutColl = "testCollection"
)

type processorTestSuite struct {
	suite.Suite
	proc   Processor
	nutsDb *nutsdb.DB
}

func TestStartConnectionSuit(t *testing.T) {
	suite.Run(t, new(processorTestSuite))
}

func (cts *processorTestSuite) SetupTest() {
	opt := nutsdb.DefaultOptions
	opt.Dir = dbDir
	nutsDb, err := nutsdb.Open(opt)
	if err != nil {
		panic(err)
	}
	proc := NewSimpleProcessor(nutsDb, Name(nutColl), filters.FilterText)
	cts.proc = proc
	cts.nutsDb = nutsDb
}

func (cts *processorTestSuite) TearDownTest() {
	if err := cts.nutsDb.Close(); err != nil {
		panic(err)
	}
	if err := os.RemoveAll(dbDir); err != nil {
		panic(err)
	}
}

func (cts *processorTestSuite) TestConnection_NewConnection() {
	cts.DirExists(dbDir)
}

func (cts *processorTestSuite) TestNutsRepository_Save1() {
	saveData := []RawData{{Url: "test", Data: "data1 data2"}}
	cts.NoError(cts.proc.ProcessAndInsertString(saveData))
	if err := cts.nutsDb.View(
		func(tx *nutsdb.Tx) error {
			key := []byte("data1")
			bucket := dataPrefix + nutColl
			if e, err := tx.SMembers(bucket, key); err != nil {
				return err
			} else {
				data := make([]string, 0, len(e))
				for i := range e {
					data = append(data, string(e[i]))
				}
				cts.ElementsMatch([]string{`{"Url":"test","Pos":[0]}`}, data)
			}
			return nil
		}); err != nil {
		panic(err)
	}
}

func (cts *processorTestSuite) TestNutsRepository_Save2() {

	now := time.Now()

	saveData := []RawData{
		{
			Url:  "source1",
			Data: "data1 data2 data2",
			Source: Source{
				Date:  now,
				Title: "Test Title",
			},
		},
		{
			Url:  "source2",
			Data: "data3 data2",
			Source: Source{
				Date:  now,
				Title: "Test Second Title",
			},
		},
	}
	cts.NoError(cts.proc.ProcessAndInsertString(saveData))
	if err := cts.nutsDb.View(
		func(tx *nutsdb.Tx) error {
			key := []byte("data2")
			bucket := dataPrefix + nutColl
			if e, err := tx.SMembers(bucket, key); err != nil {
				return err
			} else {
				data := make([]string, 0, len(e))
				for i := range e {
					data = append(data, string(e[i]))
				}
				cts.ElementsMatch([]string{`{"Url":"source1","Pos":[1,2]}`, `{"Url":"source2","Pos":[1]}`}, data)
			}
			return nil
		}); err != nil {
		panic(err)
	}

	if err := cts.nutsDb.View(
		func(tx *nutsdb.Tx) error {
			key := []byte("source1")
			bucket := sourceBucket
			if e, err := tx.Get(bucket, key); err != nil {
				return err
			} else {
				data, err := json.Marshal(&Source{
					Date:  now,
					Title: "Test Title",
				})
				if err != nil {
					return err
				}
				cts.JSONEq(string(data), string(e.Value))
			}
			return nil
		}); err != nil {
		panic(err)
	}

}
