package repo_test

import (
	"log"
	"path"
	"testing"

	"github.com/bongofriend/bookplayer/backend/lib/config"
	"github.com/bongofriend/bookplayer/backend/lib/data/repo"
)

func TestApplyMigrations(t *testing.T) {
	tmpDir := t.TempDir()
	testConfig := config.DbConfig{
		Path:       path.Join(tmpDir, "test.db"),
		DriverName: "sqlite3",
	}

	if err := repo.ApplyDatabaseMigrations(testConfig); err != nil {
		log.Fatal(err)
	}
}
