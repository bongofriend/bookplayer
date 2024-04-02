package metadataextractor_test

import (
	"context"
	"log"
	"sync"
	"testing"
	"time"

	"github.com/bongofriend/bookplayer/backend/lib/models"
	"github.com/bongofriend/bookplayer/backend/lib/processing"
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
	pathChan := make(chan processing.AudiobookDiscoveryResult)
	context, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	wg := sync.WaitGroup{}
	done := make(chan bool)
	var audiobook *models.Audiobook
	go func() {
		select {
		case <-context.Done():
			done <- false
			close(done)
			return
		case data := <-extractor.MetadataChan:
			audiobook = &data.Audiobook
			done <- true
			close(done)
		}
	}()

	extractor.Process(context, &wg, pathChan)
	pathChan <- testfilePath
	<-done

	cancel()
	wg.Wait()

	if audiobook == nil {
		log.Fatal("No data received from MetadataExtractor")
	}
}
