package chaptersplitter_test

import (
	"context"
	"encoding/json"
	"log"
	"os"
	"path"
	"path/filepath"
	"sync"
	"testing"

	"github.com/bongofriend/bookplayer/backend/lib"
	"github.com/bongofriend/bookplayer/backend/lib/processing/chaptersplitter"
)

const testJson = `{"Title": "The Art of War (Unabridged)","Author":"Sun Tzu","Narrator":"Aidan Gillen","Description":"The 13 chapters of The Art of War, each devoted to one aspect of warfare, were compiled by the high-ranking Chinese military general, strategist, and philosopher Sun-Tzu....","Genre":"Audiobook","Duration":4077.439,"Chapters":[{"Title":" Opening Credits ","StartTime":0,"EndTime":20.526,"Start":0,"End":20526},{"Title":" 1. Laying Plans ","StartTime":20.526,"EndTime":276.921,"Start":20526,"End":276921},{"Title":" 2. Waging War ","StartTime":276.921,"EndTime":509.213,"Start":276921,"End":509213},{"Title":" 3. Attack by Stratagem ","StartTime":509.213,"EndTime":766.909,"Start":509213,"End":766909},{"Title":" 4. Tactical Dispositions ","StartTime":766.909,"EndTime":964.372,"Start":766909,"End":964372},{"Title":" 5. Energy ","StartTime":964.372,"EndTime":1216.68,"Start":964372,"End":1216680},{"Title":" 6. Weak Points and Strong ","StartTime":1216.68,"EndTime":1590.893,"Start":1216680,"End":1590893},{"Title":" 7. Manœuvring ","StartTime":1590.893,"EndTime":1923.774,"Start":1590893,"End":1923774},{"Title":" 8. Variation in Tactics ","StartTime":1923.774,"EndTime":2086.267,"Start":1923774,"End":2086267},{"Title":" 9. The Army on the March ","StartTime":2086.267,"EndTime":2516.069,"Start":2086267,"End":2516069},{"Title":" 10. Terrain ","StartTime":2516.069,"EndTime":2870.314,"Start":2516069,"End":2870314},{"Title":" 11. The Nine Situations ","StartTime":2870.314,"EndTime":3546.805,"Start":2870314,"End":3546805},{"Title":" 12. The Attack By Fire ","StartTime":3546.805,"EndTime":3741.296,"Start":3546805,"End":3741296},{"Title":" 13. The Use of Spies ","StartTime":3741.296,"EndTime":4077.429,"Start":3741296,"End":4077429}],"FilePath":"/home/memi/projects/bookplayer/data/test.m4b"}`

func TestChapterSplitter(t *testing.T) {
	audiobook := lib.Audiobook{}
	if err := json.Unmarshal([]byte(testJson), &audiobook); err != nil {
		t.Fatal(err)
	}

	chaptersplitter, err := chaptersplitter.NewChapterSplitter()
	if err != nil {
		t.Fatal(err)
	}
	inputChan := make(chan lib.Audiobook, 1)
	done := make(chan bool)
	config := lib.ProcessedAudiobooksConfig{
		ProcessedPath: t.TempDir(),
	}
	wg := sync.WaitGroup{}
	context, cancel := context.WithCancel(context.Background())
	var output *lib.Audiobook

	go func() {
		select {
		case <-context.Done():
			done <- false
			close(done)
			return
		case a := <-chaptersplitter.OutputChan:
			output = &a
			done <- true
			close(done)
			return
		}
	}()

	inputChan <- audiobook
	chaptersplitter.Start(context, &wg, inputChan, config)
	<-done
	close(inputChan)
	cancel()
	wg.Wait()

	if output == nil {
		log.Fatal("no output received")
	}

	f := path.Join(config.ProcessedPath, output.Title)
	_, err = os.Stat(f)
	if os.IsNotExist(err) {
		log.Fatalf("%s was not found", f)
	}

	chapterFiles, err := filepath.Glob(path.Join(f, "*.m4b"))
	if err != nil {
		log.Fatal(err)
	}

	if len(chapterFiles) != len(audiobook.Chapters) {
		log.Fatalf("Expected: %d chapter files, Found: %d chapter files", len(audiobook.Chapters), len(chapterFiles))
	}

}
