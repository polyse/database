// Package db is used for low-level operations related to communication with the database.
package collection

import (
	"github.com/rs/zerolog"
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
	db      *nutsdb.DB
	colName string
	l       zerolog.Logger
}

// Config describes the basic database configuration.
type Config struct {
	File string
}

// CollectionName sets the name for the collection to be contained in the repository.
type Name string

// NewNutRepo function-constructor to NutsRepository.
func NewNutRepo(colName Name, db *nutsdb.DB) *NutsRepository {
	l := log.With().Str("collection name", string(colName)).Logger()
	l.Debug().Msg("initialize repository")
	return &NutsRepository{db: db, colName: string(colName), l: l}
}

// GetCollectionName returns the name of the collection specified for this repository.
func (nr *NutsRepository) GetCollectionName() string {
	return nr.colName
}

// Save saves data to the collection of this repository.
func (nr *NutsRepository) Save(ent map[string][]string) error {

	nr.l.Debug().Interface("data", ent).Msg("start inserting data")

	return nr.db.Update(func(tx *nutsdb.Tx) error {
		for i := range ent {
			vals := ent[i]
			data := make([][]byte, 0, len(vals))
			for j := range vals {
				data = append(data, []byte(vals[j]))
			}
			if err := tx.SAdd(nr.colName, []byte(i), data...); err != nil {
				nr.l.Err(err).
					Str("key", i).
					Strs("values", ent[i]).
					Msg("can not SADD to database")
				if errC := tx.Rollback(); errC != nil {
					nr.l.Err(err).Msg("can not rollback transaction")
				}
				return err
			}
		}
		return nil
	})

}
