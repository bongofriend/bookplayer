package processing_test

import (
	"context"
	"encoding/json"
	"log"
	"os"
	"path"
	"testing"

	"github.com/bongofriend/bookplayer/backend/lib/config"
	"github.com/bongofriend/bookplayer/backend/lib/models"
	"github.com/bongofriend/bookplayer/backend/lib/processing"
)

const testJson = "{\"Title\":\"The Art of War (Unabridged)\",\"Author\":\"Sun Tzu\",\"Narrator\":\"Aidan Gillen\",\"Description\":\"The 13 chapters of The Art of War, each devoted to one aspect of warfare, were compiled by the high-ranking Chinese military general, strategist, and philosopher Sun-Tzu....\",\"Genre\":\"Audiobook\",\"Duration\":4077.439,\"Chapters\":[{\"Title\":\" Opening Credits \",\"StartTime\":0,\"EndTime\":20.526,\"Start\":0,\"End\":20526,\"Numbering\":0},{\"Title\":\" 1. Laying Plans \",\"StartTime\":20.526,\"EndTime\":276.921,\"Start\":20526,\"End\":276921,\"Numbering\":1},{\"Title\":\" 2. Waging War \",\"StartTime\":276.921,\"EndTime\":509.213,\"Start\":276921,\"End\":509213,\"Numbering\":2},{\"Title\":\" 3. Attack by Stratagem \",\"StartTime\":509.213,\"EndTime\":766.909,\"Start\":509213,\"End\":766909,\"Numbering\":3},{\"Title\":\" 4. Tactical Dispositions \",\"StartTime\":766.909,\"EndTime\":964.372,\"Start\":766909,\"End\":964372,\"Numbering\":4},{\"Title\":\" 5. Energy \",\"StartTime\":964.372,\"EndTime\":1216.68,\"Start\":964372,\"End\":1216680,\"Numbering\":5},{\"Title\":\" 6. Weak Points and Strong \",\"StartTime\":1216.68,\"EndTime\":1590.893,\"Start\":1216680,\"End\":1590893,\"Numbering\":6},{\"Title\":\" 7. Manœuvring \",\"StartTime\":1590.893,\"EndTime\":1923.774,\"Start\":1590893,\"End\":1923774,\"Numbering\":7},{\"Title\":\" 8. Variation in Tactics \",\"StartTime\":1923.774,\"EndTime\":2086.267,\"Start\":1923774,\"End\":2086267,\"Numbering\":8},{\"Title\":\" 9. The Army on the March \",\"StartTime\":2086.267,\"EndTime\":2516.069,\"Start\":2086267,\"End\":2516069,\"Numbering\":9},{\"Title\":\" 10. Terrain \",\"StartTime\":2516.069,\"EndTime\":2870.314,\"Start\":2516069,\"End\":2870314,\"Numbering\":10},{\"Title\":\" 11. The Nine Situations \",\"StartTime\":2870.314,\"EndTime\":3546.805,\"Start\":2870314,\"End\":3546805,\"Numbering\":11},{\"Title\":\" 12. The Attack By Fire \",\"StartTime\":3546.805,\"EndTime\":3741.296,\"Start\":3546805,\"End\":3741296,\"Numbering\":12},{\"Title\":\" 13. The Use of Spies \",\"StartTime\":3741.296,\"EndTime\":4077.429,\"Start\":3741296,\"End\":4077429,\"Numbering\":13}]}"
const testFilePath = "/home/memi/projects/bookplayer/data/test.m4b"

func TestChapterSplitter(t *testing.T) {
	audiobook := models.Audiobook{}
	if err := json.Unmarshal([]byte(testJson), &audiobook); err != nil {
		t.Fatal(err)
	}
	config := config.Config{
		ApplicationDirectory: t.TempDir(),
	}
	handler, err := processing.NewChapterSplitter(config)
	if err != nil {
		t.Fatal(err)
	}
	chapterSplitter := processing.NewPipelineStage[processing.AudiobookMetadataResult, models.AudiobookProcessed](handler)
	doneConsumer := make(chan bool)
	errChan := make(chan error)

	context, cancel := context.WithCancel(context.Background())
	var result *models.AudiobookProcessed

	go func() {
		select {
		case <-context.Done():
			doneConsumer <- false
			close(doneConsumer)
			return
		case a := <-chapterSplitter.OutputChan:
			result = &a
			doneConsumer <- true
			close(doneConsumer)
			return
		}
	}()

	go chapterSplitter.Start(context, errChan)
	chapterSplitter.InputChan <- processing.AudiobookMetadataResult{
		Audiobook: audiobook,
		FilePath:  testFilePath,
	}

	<-doneConsumer
	cancel()
	<-chapterSplitter.DoneChan

	if result == nil {
		log.Fatal("no output received")
	}

	f := path.Join(config.ApplicationDirectory, result.Title)
	_, err = os.Stat(f)
	if os.IsNotExist(err) {
		log.Fatalf("%s was not found", f)
	}

	if len(result.ProcessedChapters) != len(audiobook.Chapters) {
		log.Fatalf("Expected: %d chapter files, Found: %d chapter files", len(audiobook.Chapters), len(result.ProcessedChapters))
	}

}
