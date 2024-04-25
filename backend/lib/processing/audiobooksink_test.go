package processing_test

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"

	"github.com/bongofriend/bookplayer/backend/lib/models"
	"github.com/bongofriend/bookplayer/backend/lib/processing"
)

const testAudiobook string = `{"Title":"The Art of War (Unabridged)","Author":"Sun Tzu","Narrator":"Aidan Gillen","Description":"The 13 chapters of The Art of War, each devoted to one aspect of warfare, were compiled by the high-ranking Chinese military general, strategist, and philosopher Sun-Tzu....","Genre":"Audiobook","Duration":4077.439,"FilePath":"/home/memi/projects/bookplayer/data/test.m4b","ProcessedChapters":[{"Title":" Opening Credits ","StartTime":0,"EndTime":20.526,"Numbering":0,"FilePath":"/tmp/TestChapterSplitter2311931691/001/The Art of War (Unabridged)/0.m4b"},{"Title":" 1. Laying Plans ","StartTime":20.526,"EndTime":276.921,"Numbering":1,"FilePath":"/tmp/TestChapterSplitter2311931691/001/The Art of War (Unabridged)/1.m4b"},{"Title":" 2. Waging War ","StartTime":276.921,"EndTime":509.213,"Numbering":2,"FilePath":"/tmp/TestChapterSplitter2311931691/001/The Art of War (Unabridged)/2.m4b"},{"Title":" 3. Attack by Stratagem ","StartTime":509.213,"EndTime":766.909,"Numbering":3,"FilePath":"/tmp/TestChapterSplitter2311931691/001/The Art of War (Unabridged)/3.m4b"},{"Title":" 4. Tactical Dispositions ","StartTime":766.909,"EndTime":964.372,"Numbering":4,"FilePath":"/tmp/TestChapterSplitter2311931691/001/The Art of War (Unabridged)/4.m4b"},{"Title":" 5. Energy ","StartTime":964.372,"EndTime":1216.68,"Numbering":5,"FilePath":"/tmp/TestChapterSplitter2311931691/001/The Art of War (Unabridged)/5.m4b"},{"Title":" 6. Weak Points and Strong ","StartTime":1216.68,"EndTime":1590.893,"Numbering":6,"FilePath":"/tmp/TestChapterSplitter2311931691/001/The Art of War (Unabridged)/6.m4b"},{"Title":" 7. Manœuvring ","StartTime":1590.893,"EndTime":1923.774,"Numbering":7,"FilePath":"/tmp/TestChapterSplitter2311931691/001/The Art of War (Unabridged)/7.m4b"},{"Title":" 8. Variation in Tactics ","StartTime":1923.774,"EndTime":2086.267,"Numbering":8,"FilePath":"/tmp/TestChapterSplitter2311931691/001/The Art of War (Unabridged)/8.m4b"},{"Title":" 9. The Army on the March ","StartTime":2086.267,"EndTime":2516.069,"Numbering":9,"FilePath":"/tmp/TestChapterSplitter2311931691/001/The Art of War (Unabridged)/9.m4b"},{"Title":" 10. Terrain ","StartTime":2516.069,"EndTime":2870.314,"Numbering":10,"FilePath":"/tmp/TestChapterSplitter2311931691/001/The Art of War (Unabridged)/10.m4b"},{"Title":" 11. The Nine Situations ","StartTime":2870.314,"EndTime":3546.805,"Numbering":11,"FilePath":"/tmp/TestChapterSplitter2311931691/001/The Art of War (Unabridged)/11.m4b"},{"Title":" 12. The Attack By Fire ","StartTime":3546.805,"EndTime":3741.296,"Numbering":12,"FilePath":"/tmp/TestChapterSplitter2311931691/001/The Art of War (Unabridged)/12.m4b"},{"Title":" 13. The Use of Spies ","StartTime":3741.296,"EndTime":4077.429,"Numbering":13,"FilePath":"/tmp/TestChapterSplitter2311931691/001/The Art of War (Unabridged)/13.m4b"}]}`

type audiobookMockRepository struct {
	currentId int64
	data      map[int64]models.AudiobookProcessed
}

func (a audiobookMockRepository) GetAudiobookById(context context.Context, id int64) (*models.AudiobookProcessed, error) {
	audiobook, ok := a.data[id]
	if !ok {
		return nil, fmt.Errorf("Audiobook with Id %d not found", id)
	}
	return &audiobook, nil
}

func (a *audiobookMockRepository) InsertAudiobook(context context.Context, audiobook models.AudiobookProcessed) (int64, error) {
	a.currentId++
	a.data[a.currentId] = audiobook
	return a.currentId, nil
}

func TestAudiobookSink(t *testing.T) {
	mockRepo := audiobookMockRepository{
		currentId: 0,
		data:      map[int64]models.AudiobookProcessed{},
	}
	audiobook := models.AudiobookProcessed{}
	if err := json.Unmarshal([]byte(testAudiobook), &audiobook); err != nil {
		t.Fatal(err)
	}
	sinkHandler := processing.NewAudiobookSink(&mockRepo)
	sink := processing.NewPipelineStage(sinkHandler)
	context, cancel := context.WithCancel(context.Background())

	doneConsumerChan := make(chan struct{})
	outputReceived := false

	go func() {
		defer func() {
			doneConsumerChan <- struct{}{}
			close(doneConsumerChan)
		}()
		select {
		case <-doneConsumerChan:
			return
		case <-sink.OutputChan:
			outputReceived = true
		}
	}()

	go sink.Start(context)
	sink.InputChan <- audiobook

	<-doneConsumerChan
	cancel()
	<-sink.DoneChan

	if !outputReceived {
		t.Fatal("No output received")
	}

	if len(mockRepo.data) == 0 {
		t.Fatal("No audiobook was inserted")
	}
}
