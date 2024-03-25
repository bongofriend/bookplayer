package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"

	config "github.com/bongofriend/backend/lib"
	directorywatcher "github.com/bongofriend/backend/lib/processing"
)

func main() {
	config, err := config.GetConfig(config.Dev)
	if err != nil {
		log.Fatal(err)
	}
	context, cancel := context.WithCancel(context.Background())

	wg := sync.WaitGroup{}
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGTERM, syscall.SIGINT)

	wg.Add(1)
	watcher := directorywatcher.DirectoryWatcher{}
	watcher.Start(context, &wg, config.Audiobooks)

	<-sigChan
	cancel()
	wg.Wait()
}
