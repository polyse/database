package db

import (
	"github.com/rs/zerolog/log"
	"github.com/xujiajun/nutsdb"
)

// NutConnection structure describing connecting to NutDB.
type NutConnection struct {
	db *nutsdb.DB
}

// NewNutConnection function-constructor to NewNutConnection.
func NewNutConnection(cfg Config) (*NutConnection, func(), error) {
	opt := nutsdb.DefaultOptions
	opt.Dir = string(cfg)
	db, err := nutsdb.Open(opt)
	if err != nil {
		return nil, nil, err
	}
	conn := &NutConnection{db: db}
	return conn, func() {
		if err = conn.Close(); err != nil {
			log.Err(err).Msg("can not close database connection")
		}
	}, nil
}

// Close merge database files and close connection to NutDB.
func (c *NutConnection) Close() error {
	log.Debug().Msg("start closing database connection")
	if err := c.db.Merge(); err != nil {
		log.Err(err).Msg("can not merge database")
	}
	return c.db.Close()
}
