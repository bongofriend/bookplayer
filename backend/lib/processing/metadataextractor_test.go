package processing_test

import (
	"context"
	"log"
	"testing"
	"time"

	"github.com/bongofriend/bookplayer/backend/lib/models"
	"github.com/bongofriend/bookplayer/backend/lib/processing"
)

const (
	testfilePath = "/home/memi/projects/bookplayer/data/test.m4b"
)

func TestNewMetadataExtractor(t *testing.T) {
	if _, err := processing.NewMetadataExtractor(); err != nil {
		t.Fatal(err)
	}
}

func TestMetaDataExtractorProcess(t *testing.T) {
	extractorHandler, _ := processing.NewMetadataExtractor()
	extractor := processing.NewPipelineStage[string, processing.AudiobookMetadataResult](extractorHandler)
	context, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	doneConsumer := make(chan struct{})
	errChan := make(chan error)

	go func() {
		select {
		case <-context.Done():
			return
		case err := <-errChan:
			cancel()
			log.Println(err)
			return
		}
	}()

	go extractor.Start(context, errChan)
	extractor.InputChan <- testfilePath

	var audiobook *models.Audiobook
	go func() {
		defer func() {
			doneConsumer <- struct{}{}
		}()
		select {
		case <-context.Done():
			return
		case data := <-extractor.OutputChan:
			audiobook = &data.Audiobook
			return
		}
	}()

	<-doneConsumer
	cancel()
	<-extractor.DoneChan

	if audiobook == nil {
		log.Fatal("No data received from MetadataExtractor")
	}
}
