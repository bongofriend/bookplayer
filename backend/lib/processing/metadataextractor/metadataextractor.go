package metadataextractor

import (
	"context"
	"errors"
	"log"
	"os/exec"
	"sync"
)

type MetadataExtractor struct {
}

func NewMetadataExtractor() (MetadataExtractor, error) {
	if !ffmpegIsAvailable() {
		return MetadataExtractor{}, errors.New("ffmpeg is not installed or found")
	}
	return MetadataExtractor{}, nil
}

// TODO
func (m MetadataExtractor) extractMetadata(path string) {

}

func (m MetadataExtractor) Process(ctx context.Context, wg *sync.WaitGroup, pathChan <-chan string) {
	wg.Add(1)
	go func() {
		defer wg.Done()
		for {
			select {
			case <-ctx.Done():
				log.Println("Shutting down MetadataExtractor")
				return
			case path := <-pathChan:
				m.extractMetadata(path)
			}
		}
	}()
}

func ffmpegIsAvailable() bool {
	_, err := exec.LookPath("ffmpeg")
	return err == nil
}
