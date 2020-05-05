package db

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

type repositoryTestSuite struct {
	suite.Suite
	repo Repository
	con  *NutConnection
}

func TestStartConnectionSuit(t *testing.T) {
	suite.Run(t, new(repositoryTestSuite))
}

func (cts *repositoryTestSuite) SetupTest() {
	c, _, err := NewNutConnection(Config(dbDir))
	if err != nil {
		panic(err)
	}
	repo := NewNutRepo(CollectionName(nutColl), c)
	cts.repo = repo
	cts.con = c
}

func (cts *repositoryTestSuite) TearDownTest() {
	if err := cts.con.Close(); err != nil {
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
	saveData := map[string][]string{
		"test": {
			"data1",
			"data2",
		},
	}
	cts.NoError(cts.repo.Save(saveData))
	if err := cts.con.db.View(
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
	saveData := map[string][]string{
		"test0": {
			"data0-0",
			"data0-1",
		},
		"test1": {
			"data1-0",
		},
	}
	cts.NoError(cts.repo.Save(saveData))
	if err := cts.con.db.View(
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

	cts.NoError(cts.repo.Save(map[string][]string{
		"test1": {
			"data1-1",
		},
	}))

	if err := cts.con.db.View(
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
