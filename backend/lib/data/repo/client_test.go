package repo_test

import (
	"log"
	"path"
	"testing"

	"github.com/bongofriend/bookplayer/backend/lib/config"
	"github.com/bongofriend/bookplayer/backend/lib/data/repo"
)

func TestDbClient(t *testing.T) {
	tmpDir := t.TempDir()
	testConfig := config.DatabaseConfig{
		Path:   path.Join(tmpDir, "test.db"),
		Driver: "sqlite3",
	}

	client, err := repo.NewDbClient(testConfig)
	if err != nil {
		log.Fatal(err)
	}
	if err := client.Close(); err != nil {
		log.Fatal(err)
	}
}
