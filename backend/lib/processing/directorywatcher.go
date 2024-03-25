package directorywatcher

import (
	"context"
	"log"
	"sync"
	"time"

	config "github.com/bongofriend/backend/lib"
)

type DirectoryWatcher struct {
	PathChan chan string
}

func (d DirectoryWatcher) shutdown(c config.AudiobooksConfig) {
	log.Printf("Stopping to watch %s", c.AudibookDirectoryPath)
}

func (d DirectoryWatcher) parseDirectoryContent() {
	//TODO Implement
}

func (d DirectoryWatcher) Start(ctx context.Context, wg *sync.WaitGroup, c config.AudiobooksConfig) error {
	ticker := time.NewTicker(c.Interval)
	log.Printf("Watching %s ...", c.AudibookDirectoryPath)
	go func() {
		for {
			select {
			case <-ctx.Done():
				d.shutdown(c)
				wg.Done()
				return
			case <-ticker.C:
				d.parseDirectoryContent()
			}
		}
	}()
	return nil
}
