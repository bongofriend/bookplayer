package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/bongofriend/bookplayer/backend/lib"
	"github.com/bongofriend/bookplayer/backend/lib/processing/directorywatcher"
)

func main() {
	envPath, err := lib.GetEnvPathFromFlags()
	if err != nil {
		log.Fatal(err)
	}

	config, err := lib.ParseConfig(envPath)
	if err != nil {
		log.Fatal(err)
	}
	ctx, cancel := context.WithCancel(context.Background())

	wg := sync.WaitGroup{}
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGTERM, syscall.SIGINT)

	watcher := directorywatcher.NewDirectoryWatcher()
	watcher.Start(ctx, &wg, config.Audiobooks)

	<-sigChan
	log.Println("Shutting down")
	cancel()
	wg.Wait()
}
