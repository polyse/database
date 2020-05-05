package db

import (
	"github.com/rs/zerolog/log"
	"github.com/xujiajun/nutsdb"
)

type Connection struct {
	db *nutsdb.DB
}

func NewConnection(cfg Config) (*Connection, func(), error) {
	opt := nutsdb.DefaultOptions
	opt.Dir = string(cfg)
	db, err := nutsdb.Open(opt)
	if err != nil {
		return nil, nil, err
	}
	conn := &Connection{db: db}
	return conn, func() {
		if err = conn.Close(); err != nil {
			log.Err(err).Msg("can not close database connection")
		}
	}, nil
}

func (c *Connection) Close() error {
	log.Debug().Msg("start closing database connection")
	return c.db.Close()
}
