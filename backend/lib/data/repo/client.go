package repo

import (
	"database/sql"

	"github.com/bongofriend/bookplayer/backend/lib/config"
	"github.com/bongofriend/bookplayer/backend/lib/data/datasource"
	_ "github.com/mattn/go-sqlite3"
)

type DbClient struct {
	queries datasource.Queries
	db      *sql.DB
}

func NewDbClient(config config.DatabaseConfig) (*DbClient, error) {
	db, err := sql.Open(config.Driver, config.Path)
	if err != nil {
		return nil, err
	}
	return &DbClient{
		db:      db,
		queries: *datasource.New(db),
	}, nil
}

func (c *DbClient) Close() error {
	return c.db.Close()
}
