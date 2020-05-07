package collection

import (
	"os"
	"testing"

	"github.com/stretchr/testify/suite"
	"github.com/xujiajun/nutsdb"
)

var (
	dbDir   = "nutsdb-test"
	nutColl = "testCollection"
)

type testByte struct {
	str string
}

func (tb *testByte) GetBytes() []byte {
	return []byte(tb.str)
}

type repositoryTestSuite struct {
	suite.Suite
	repo   Repository
	nutsDb *nutsdb.DB
}

func TestStartConnectionSuit(t *testing.T) {
	suite.Run(t, new(repositoryTestSuite))
}

func (cts *repositoryTestSuite) SetupTest() {
	opt := nutsdb.DefaultOptions
	opt.Dir = dbDir
	nutsDb, err := nutsdb.Open(opt)
	if err != nil {
		panic(err)
	}
	repo := NewNutRepo(Name(nutColl), nutsDb)
	cts.repo = repo
	cts.nutsDb = nutsDb
}

func (cts *repositoryTestSuite) TearDownTest() {
	if err := cts.nutsDb.Close(); err != nil {
		panic(err)
	}
	if err := os.RemoveAll(dbDir); err != nil {
		panic(err)
	}
}

func (cts *repositoryTestSuite) TestConnection_NewConnection() {
	cts.DirExists(dbDir)
}

func (cts *repositoryTestSuite) TestNutsRepository_GetCollectionName() {
	cts.Equal(nutColl, cts.repo.GetCollectionName())
}

func (cts *repositoryTestSuite) TestNutsRepository_Save1() {
	saveData := map[string][]ByteArr{
		"test": {
			&testByte{str: "data1"},
			&testByte{str: "data2"},
		},
	}
	cts.NoError(cts.repo.Save(saveData))
	if err := cts.nutsDb.View(
		func(tx *nutsdb.Tx) error {
			key := []byte("test")
			bucket := cts.repo.GetCollectionName()
			if e, err := tx.SMembers(bucket, key); err != nil {
				return err
			} else {
				data := make([]string, 0, len(e))
				for i := range e {
					data = append(data, string(e[i]))
				}
				cts.ElementsMatch([]string{"data1", "data2"}, data)
			}
			return nil
		}); err != nil {
		panic(err)
	}
}

func (cts *repositoryTestSuite) TestNutsRepository_Save2() {
	saveData := map[string][]ByteArr{
		"test0": {
			&testByte{"data0-0"},
			&testByte{"data0-1"},
		},
		"test1": {
			&testByte{"data1-0"},
		},
	}
	cts.NoError(cts.repo.Save(saveData))
	if err := cts.nutsDb.View(
		func(tx *nutsdb.Tx) error {
			key := []byte("test1")
			bucket := cts.repo.GetCollectionName()
			if e, err := tx.SMembers(bucket, key); err != nil {
				return err
			} else {
				data := make([]string, 0, len(e))
				for i := range e {
					data = append(data, string(e[i]))
				}
				cts.ElementsMatch([]string{"data1-0"}, data)
			}
			return nil
		}); err != nil {
		panic(err)
	}

	cts.NoError(cts.repo.Save(map[string][]ByteArr{
		"test1": {
			&testByte{"data1-1"},
		},
	}))

	if err := cts.nutsDb.View(
		func(tx *nutsdb.Tx) error {
			key := []byte("test1")
			bucket := cts.repo.GetCollectionName()
			if e, err := tx.SMembers(bucket, key); err != nil {
				return err
			} else {
				data := make([]string, 0, len(e))
				for i := range e {
					data = append(data, string(e[i]))
				}
				cts.ElementsMatch([]string{"data1-0", "data1-1"}, data)
			}
			return nil
		}); err != nil {
		panic(err)
	}

}
