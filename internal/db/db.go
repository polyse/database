package db

import (
	"github.com/rs/zerolog/log"
)

type Repository interface {
	Save(ent map[string][]string) error
	GetCollectionName() string
}

type NutsRepository struct {
	c       *Connection
	colName string
}

type Config string
type CollectionName string

func NewNutRepo(colName CollectionName, c *Connection) *NutsRepository {
	return &NutsRepository{c: c, colName: string(colName)}
}

func (nr *NutsRepository) GetCollectionName() string {
	return nr.colName
}

func (nr *NutsRepository) Save(ent map[string][]string) error {
	tx, err := nr.c.db.Begin(true)
	if err != nil {
		return err
	}

	for i := range ent {
		vals := ent[i]
		data := make([][]byte, 0, len(vals))
		for j := range vals {
			data = append(data, []byte(vals[j]))
		}
		if err := tx.SAdd(nr.colName, []byte(i), data...); err != nil {
			log.Err(err).
				Str("collection name", nr.colName).
				Str("key", i).
				Strs("values", ent[i]).
				Msg("can not sadd to database")
			if errC := tx.Rollback(); errC != nil {
				log.Fatal().Err(err).Msg("can not rollback transaction")
			}
			return err
		}
	}
	if err = tx.Commit(); err != nil {
		log.Err(err).
			Str("collection name", nr.colName).
			Msg("can not commit transaction")
		if errC := tx.Rollback(); errC != nil {
			log.Fatal().Err(err).Msg("can not rollback transaction")
		}
		return err
	}
	return nil
}
