package repo

import (
	"database/sql"

	"github.com/bongofriend/bookplayer/backend/lib/config"
	"github.com/bongofriend/bookplayer/backend/lib/data/datasource"
)

type DbClient struct {
	queries datasource.Queries
	db      *sql.DB
}

func NewDbClient(config config.DbConfig) (*DbClient, error) {
	db, err := sql.Open(config.DriverName, config.Path)
	if err != nil {
		return nil, err
	}
	return &DbClient{
		db:      db,
		queries: *datasource.New(db),
	}, nil
}
