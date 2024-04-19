package audiobooksink

import (
	"context"
	"log"

	"github.com/bongofriend/bookplayer/backend/lib/data/repo"
	"github.com/bongofriend/bookplayer/backend/lib/models"
)

// TODO Implement PipelineHandler interface
type AudiobookSink struct {
	audiobookRepo repo.AudiobookRepository
}

func NewAudiobookSink(audiobookRepository repo.AudiobookRepository) AudiobookSink {
	return AudiobookSink{
		audiobookRepo: audiobookRepository,
	}
}

func (a AudiobookSink) Shutdown() {
	log.Println("Shutting down AudiobookSink")
}

func (a AudiobookSink) Handle(input models.AudiobookProcessed, outputChan chan struct{}) error {
	_, err := a.audiobookRepo.InsertAudiobook(context.Background(), input)
	outputChan <- struct{}{}
	return err
}
