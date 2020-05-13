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
	proc := NewSimpleProcessor(
		nutsDb,
		Name(nutColl),
		filters.FilterText,
		filters.StemmAndToLower,
		filters.StopWords,
	)
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
	saveData = []RawData{
		{
			Url:  "source1",
			Data: "a Data5",
			Source: Source{
				Date:  now,
				Title: "Test Title New",
			},
		},
	}
	cts.NoError(cts.proc.ProcessAndInsertString(saveData))

	if err := cts.nutsDb.View(
		func(tx *nutsdb.Tx) error {
			key := []byte("data5")
			bucket := dataPrefix + nutColl
			if e, err := tx.SMembers(bucket, key); err != nil {
				return err
			} else {
				data := make([]string, 0, len(e))
				for i := range e {
					data = append(data, string(e[i]))
				}
				cts.ElementsMatch([]string{`{"Url":"source1","Pos":[0]}`}, data)
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
					Title: "Test Title New",
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

func (cts *processorTestSuite) TestNutsRepository_Get() {
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
	res, err := cts.proc.ProcessAndGet("data2", 100, 0)
	cts.NoError(err)
	cts.ElementsMatch(res, []ResponseData{
		{
			Source: Source{
				Date:  now.Round(1 * time.Nanosecond),
				Title: "Test Title",
			},
			Url: "source1",
		},
		{
			Source: Source{
				Date:  now.Round(1 * time.Nanosecond),
				Title: "Test Second Title",
			},
			Url: "source2",
		},
	})
}

func (cts *processorTestSuite) TestNutsRepository_Get2() {
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
	res, err := cts.proc.ProcessAndGet("data3 data2", 100, 0)
	cts.NoError(err)
	cts.ElementsMatch(res, []ResponseData{
		{
			Source: Source{
				Date:  now.Round(1 * time.Nanosecond),
				Title: "Test Second Title",
			},
			Url: "source2",
		},
	})
}

func (cts *processorTestSuite) TestNutsRepository_Get3() {
	now := time.Now()

	saveData := []RawData{
		{
			Url:  "source1",
			Data: "data1 data2 data2",
			Source: Source{
				Date:  now.Add(-1 * time.Hour),
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
		{
			Url:  "source3",
			Data: "data3 data2",
			Source: Source{
				Date:  now.Add(-10 * time.Minute),
				Title: "Test Second Title",
			},
		},
	}
	cts.NoError(cts.proc.ProcessAndInsertString(saveData))
	res, err := cts.proc.ProcessAndGet("data2", 100, 0)
	cts.NoError(err)
	cts.Equal(res, []ResponseData{
		{
			Source: Source{
				Date:  now.Round(1 * time.Nanosecond),
				Title: "Test Second Title",
			},
			Url: "source2",
		},
		{
			Url: "source3",
			Source: Source{
				Date:  now.Add(-10 * time.Minute).Round(1 * time.Nanosecond),
				Title: "Test Second Title",
			},
		},
		{
			Url: "source1",
			Source: Source{
				Date:  now.Add(-1 * time.Hour).Round(1 * time.Nanosecond),
				Title: "Test Title",
			},
		},
	})
}

func (cts *processorTestSuite) TestNutsRepository_Get4() {
	now := time.Now()

	saveData := []RawData{
		{
			Url:  "source1",
			Data: "data1 data2 data2",
			Source: Source{
				Date:  now.Add(-1 * time.Hour),
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
		{
			Url:  "source3",
			Data: "data3 data2",
			Source: Source{
				Date:  now.Add(-10 * time.Minute),
				Title: "Test Second Title",
			},
		},
	}
	cts.NoError(cts.proc.ProcessAndInsertString(saveData))
	res, err := cts.proc.ProcessAndGet("data2", 1, 1)
	cts.NoError(err)
	cts.Equal(res, []ResponseData{
		{
			Url: "source3",
			Source: Source{
				Date:  now.Add(-10 * time.Minute).Round(1 * time.Nanosecond),
				Title: "Test Second Title",
			},
		},
	})
}