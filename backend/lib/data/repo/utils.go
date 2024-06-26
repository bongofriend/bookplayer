package repo

import (
	"database/sql"

	"github.com/bongofriend/bookplayer/backend/db"
	"github.com/bongofriend/bookplayer/backend/lib/config"
	"github.com/pressly/goose/v3"
)

func ApplyDatabaseMigrations(dbConfig config.DatabaseConfig) error {
	database, err := sql.Open(dbConfig.Driver, dbConfig.Path)
	if err != nil {
		return err
	}
	defer database.Close()
	goose.SetBaseFS(db.MigrationsFS)
	if err := goose.SetDialect(dbConfig.Driver); err != nil {
		return err
	}
	if err := goose.Up(database, "migrations"); err != nil {
		return err
	}
	return nil
}
