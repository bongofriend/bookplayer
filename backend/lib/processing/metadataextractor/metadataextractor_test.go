package metadataextractor_test

import (
	"context"
	"log"
	"testing"
	"time"

	"github.com/bongofriend/bookplayer/backend/lib/models"
	"github.com/bongofriend/bookplayer/backend/lib/processing/metadataextractor"
)

const (
	testfilePath = "/home/memi/projects/bookplayer/data/test.m4b"
)

func TestNewMetadataExtractor(t *testing.T) {
	if _, err := metadataextractor.NewMetadataExtractor(); err != nil {
		t.Fatal(err)
	}
}

func TestMetaDataExtractorProcess(t *testing.T) {
	extractor, _ := metadataextractor.NewMetadataExtractor()
	pathChan := make(chan string)
	context, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	doneExtractor := make(chan struct{})
	doneConsumer := make(chan struct{})

	extractor.Start(context, pathChan, doneExtractor)
	pathChan <- testfilePath

	var audiobook *models.Audiobook
	go func() {
		defer func() {
			doneConsumer <- struct{}{}
		}()
		output, err := extractor.Output()
		if err != nil {
			log.Fatal(err)
		}
		select {
		case <-context.Done():
			return
		case data := <-output:
			audiobook = &data.Audiobook
			return
		}
	}()

	<-doneConsumer
	cancel()
	<-doneExtractor

	if audiobook == nil {
		log.Fatal("No data received from MetadataExtractor")
	}
}
