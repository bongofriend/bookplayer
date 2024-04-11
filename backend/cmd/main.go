package main

import (
	"database/sql"
	"flag"
	"log"
	"os/exec"

	"github.com/bongofriend/bookplayer/backend/lib/config"
	_ "github.com/mattn/go-sqlite3"
	"github.com/pressly/goose/v3"
)

type Args struct {
	configPath    *string
	migrationName *string
	up            *bool
	down          *bool
	compileSqlc   *string
}

func main() {
	args := parseFlags()
	config, err := config.ParseConfig("/home/memi/projects/bookplayer/backend/dev.yml")
	if err != nil {
		log.Fatal(err)
	}
	switch {
	case len(*args.migrationName) > 0:
		generateMigration(config.Db, args)
	case args.up != nil && *args.up:
		if *args.down {
			log.Fatal("Only up or down options allowed")
		}
		applyMigrations(config.Db)

	case args.down != nil && *args.down:
		if *args.up {
			log.Fatal("Only up or down options allowed")
		}
		teardownMigration(config.Db)
	case len(*args.compileSqlc) > 0:
		runSqlc(args)
	}
}

func parseFlags() Args {
	configPath := flag.String("configPath", "", "path to configuration file")
	migrate := flag.String("migrate", "", "name of migration")
	up := flag.Bool("up", false, "apply migrations to database")
	down := flag.Bool("down", false, "tear down applied migrations")
	compileSqlc := flag.String("compileSqlc", "", "Run sqlc to generate Go code from SQL statements.")
	flag.Parse()

	return Args{
		configPath:    configPath,
		migrationName: migrate,
		up:            up,
		down:          down,
		compileSqlc:   compileSqlc,
	}
}

func generateMigration(dbConfig config.DbConfig, args Args) {
	database := connectToDatabase(dbConfig)
	defer database.Close()
	if err := goose.SetDialect(dbConfig.DriverName); err != nil {
		log.Fatal(err)
	}
	if err := goose.Create(database, dbConfig.Migrations, *args.migrationName, "sql"); err != nil {
		log.Fatal(err)
	}
}

func connectToDatabase(dbConfig config.DbConfig) *sql.DB {
	database, err := sql.Open(dbConfig.DriverName, dbConfig.Path)
	if err != nil {
		log.Fatal(err)
	}
	if err := goose.SetDialect(dbConfig.DriverName); err != nil {
		log.Fatal(err)
	}
	return database
}

func applyMigrations(dbConfig config.DbConfig) {
	database := connectToDatabase(dbConfig)
	defer database.Close()
	if err := goose.SetDialect(dbConfig.DriverName); err != nil {
		log.Fatal(err)
	}
	if err := goose.Up(database, dbConfig.Migrations); err != nil {
		log.Fatal(err)
	}
}

func teardownMigration(dbConfig config.DbConfig) {
	database := connectToDatabase(dbConfig)
	defer database.Close()
	if err := goose.SetDialect(dbConfig.DriverName); err != nil {
		log.Fatal(err)
	}
	if err := goose.Down(database, dbConfig.Migrations); err != nil {
		log.Fatal(err)
	}
}

func runSqlc(args Args) {
	if _, err := exec.LookPath("sqlc"); err != nil {
		log.Fatal("sqlc not installed or found")
	}
	command := exec.Command("sqlc", "generate", "--file", *args.compileSqlc)
	if err := command.Run(); err != nil {
		log.Fatal(err)
	} else {
		log.Println("Succesfully compiled SQL to Go")
	}
}
