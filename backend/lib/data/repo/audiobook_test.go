package repo_test

import (
	"context"
	"encoding/json"
	"path"
	"testing"

	"github.com/bongofriend/bookplayer/backend/lib/config"
	"github.com/bongofriend/bookplayer/backend/lib/data/repo"
	"github.com/bongofriend/bookplayer/backend/lib/models"
)

const (
	audiobook = `{"Title":"The Art of War (Unabridged)","Author":"Sun Tzu","Narrator":"Aidan Gillen","Description":"The 13 chapters of The Art of War, each devoted to one aspect of warfare, were compiled by the high-ranking Chinese military general, strategist, and philosopher Sun-Tzu....","Genre":"Audiobook","Duration":4077.439,"FilePath":"/home/memi/projects/bookplayer/data/test.m4b","ProcessedChapters":[{"Title":" Opening Credits ","StartTime":0,"EndTime":20.526,"Start":0,"End":20526,"Numbering":0,"FilePath":"/tmp/TestChapterSplitter3877723229/001/The Art of War (Unabridged)/0.m4b"},{"Title":" 1. Laying Plans ","StartTime":20.526,"EndTime":276.921,"Start":20526,"End":276921,"Numbering":1,"FilePath":"/tmp/TestChapterSplitter3877723229/001/The Art of War (Unabridged)/1.m4b"},{"Title":" 2. Waging War ","StartTime":276.921,"EndTime":509.213,"Start":276921,"End":509213,"Numbering":2,"FilePath":"/tmp/TestChapterSplitter3877723229/001/The Art of War (Unabridged)/2.m4b"},{"Title":" 3. Attack by Stratagem ","StartTime":509.213,"EndTime":766.909,"Start":509213,"End":766909,"Numbering":3,"FilePath":"/tmp/TestChapterSplitter3877723229/001/The Art of War (Unabridged)/3.m4b"},{"Title":" 4. Tactical Dispositions ","StartTime":766.909,"EndTime":964.372,"Start":766909,"End":964372,"Numbering":4,"FilePath":"/tmp/TestChapterSplitter3877723229/001/The Art of War (Unabridged)/4.m4b"},{"Title":" 5. Energy ","StartTime":964.372,"EndTime":1216.68,"Start":964372,"End":1216680,"Numbering":5,"FilePath":"/tmp/TestChapterSplitter3877723229/001/The Art of War (Unabridged)/5.m4b"},{"Title":" 6. Weak Points and Strong ","StartTime":1216.68,"EndTime":1590.893,"Start":1216680,"End":1590893,"Numbering":6,"FilePath":"/tmp/TestChapterSplitter3877723229/001/The Art of War (Unabridged)/6.m4b"},{"Title":" 7. Manœuvring ","StartTime":1590.893,"EndTime":1923.774,"Start":1590893,"End":1923774,"Numbering":7,"FilePath":"/tmp/TestChapterSplitter3877723229/001/The Art of War (Unabridged)/7.m4b"},{"Title":" 8. Variation in Tactics ","StartTime":1923.774,"EndTime":2086.267,"Start":1923774,"End":2086267,"Numbering":8,"FilePath":"/tmp/TestChapterSplitter3877723229/001/The Art of War (Unabridged)/8.m4b"},{"Title":" 9. The Army on the March ","StartTime":2086.267,"EndTime":2516.069,"Start":2086267,"End":2516069,"Numbering":9,"FilePath":"/tmp/TestChapterSplitter3877723229/001/The Art of War (Unabridged)/9.m4b"},{"Title":" 10. Terrain ","StartTime":2516.069,"EndTime":2870.314,"Start":2516069,"End":2870314,"Numbering":10,"FilePath":"/tmp/TestChapterSplitter3877723229/001/The Art of War (Unabridged)/10.m4b"},{"Title":" 11. The Nine Situations ","StartTime":2870.314,"EndTime":3546.805,"Start":2870314,"End":3546805,"Numbering":11,"FilePath":"/tmp/TestChapterSplitter3877723229/001/The Art of War (Unabridged)/11.m4b"},{"Title":" 12. The Attack By Fire ","StartTime":3546.805,"EndTime":3741.296,"Start":3546805,"End":3741296,"Numbering":12,"FilePath":"/tmp/TestChapterSplitter3877723229/001/The Art of War (Unabridged)/12.m4b"},{"Title":" 13. The Use of Spies ","StartTime":3741.296,"EndTime":4077.429,"Start":3741296,"End":4077429,"Numbering":13,"FilePath":"/tmp/TestChapterSplitter3877723229/001/The Art of War (Unabridged)/13.m4b"}]}`
)

func TestAudiobookRepository(t *testing.T) {
	t.Run("should insert Audiobook", func(t *testing.T) {
		config := prepareDatabase(t)
		client, err := repo.NewDbClient(config)
		if err != nil {
			t.Fatal(err)
		}
		context := context.Background()
		audiobookRepo := repo.NewAudiobookRepository(client)
		model := getAudiobookModel()
		if _, err := audiobookRepo.InsertAudiobook(context, *model); err != nil {
			t.Fatal(err)
		}
	})
	t.Run("should fetch inserted Audiobook", func(t *testing.T) {
		config := prepareDatabase(t)
		client, err := repo.NewDbClient(config)
		if err != nil {
			t.Fatal(err)
		}
		context := context.Background()
		audiobookRepo := repo.NewAudiobookRepository(client)
		model := getAudiobookModel()
		id, err := audiobookRepo.InsertAudiobook(context, *model)
		if err != nil {
			t.Fatal(err)
		}

		fetchedAudiobook, err := audiobookRepo.GetAudiobookById(context, id)
		if err != nil {
			t.Fatal(err)
		}
		if fetchedAudiobook == nil {
			t.Fatalf("could not retrieve Audiobook with id %d from database", id)
		}
	})

}

func prepareDatabase(t *testing.T) config.DatabaseConfig {
	tmpDir := t.TempDir()
	testConfig := config.DatabaseConfig{
		Path:   path.Join(tmpDir, "test.db"),
		Driver: "sqlite3",
	}

	if err := repo.ApplyDatabaseMigrations(testConfig); err != nil {
		t.Fatal(err)
	}
	return testConfig
}

func getAudiobookModel() *models.AudiobookProcessed {
	raw := json.RawMessage(audiobook)
	var model models.AudiobookProcessed
	json.Unmarshal(raw, &model)
	return &model
}
