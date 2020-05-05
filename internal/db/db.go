// Package db is used for low-level operations related to communication with the database.
package db

import (
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
	"github.com/xujiajun/nutsdb"
)

// Repository interface describes the basic methods for obtaining and modifying data in a database.
type Repository interface {
	Save(ent map[string][]string) error
	GetCollectionName() string
}

// NutsRepository is repository interface implementation for the NutsDB database.
type NutsRepository struct {
	c       *NutConnection
	colName string
}

// Config describes the basic database configuration.
type Config string

// CollectionName sets the name for the collection to be contained in the repository.
type CollectionName string

// NewNutRepo function-constructor to NutsRepository.
func NewNutRepo(colName CollectionName, c *NutConnection) *NutsRepository {
	return &NutsRepository{c: c, colName: string(colName)}
}

// GetCollectionName returns the name of the collection specified for this repository.
func (nr *NutsRepository) GetCollectionName() string {
	return nr.colName
}

// Save saves data to the collection of this repository.
func (nr *NutsRepository) Save(ent map[string][]string) error {
	tx, err := nr.c.db.Begin(true)
	if err != nil {
		return err
	}

	log.Debug().Interface("data", ent).Msg("start inserting data")

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
				Msg("can not SADD to database")
			if errC := tx.Rollback(); errC != nil {
				log.Err(err).Msg("can not rollback transaction")
			}
			return err
		}
	}
	return commitTransaction(tx)
}

func commitTransaction(tx *nutsdb.Tx) error {
	if err := tx.Commit(); err != nil {
		if errC := tx.Rollback(); errC != nil {
			log.Err(err).Msg("can not rollback transaction")
		}
		return errors.Wrap(err, "can not commit transaction")
	}
	log.Debug().Msg("transaction committed")
	return nil
}
