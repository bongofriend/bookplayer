package metadataextractor_test

import (
	"context"
	"log"
	"sync"
	"testing"
	"time"

	"github.com/bongofriend/bookplayer/backend/lib/processing/metadataextractor"
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
	wg := sync.WaitGroup{}
	done := make(chan bool)

	testfilePath := "/home/memi/projects/bookplayer/data/test.m4b"

	go func() {
		select {
		case <-context.Done():
			done <- false
			close(done)
			return
		case data := <-extractor.MetadataChan:
			log.Printf("%+v", data)
			done <- true
			close(done)
		}
	}()

	extractor.Process(context, &wg, pathChan)
	pathChan <- testfilePath
	dataReceived := <-done

	cancel()
	wg.Wait()

	if !dataReceived {
		log.Fatal("No data received from MetadataExtractor")
	}

}
