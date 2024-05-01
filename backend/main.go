package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/bongofriend/bookplayer/backend/lib/config"
	"github.com/bongofriend/bookplayer/backend/lib/data/repo"
	"github.com/bongofriend/bookplayer/backend/lib/processing"
	_ "github.com/mattn/go-sqlite3"
)

func main() {
	/*envPath, err := config.GetEnvPathFromFlags()
	if err != nil {
		log.Fatal(err)
	}*/
	configPath := "./config.json"

	config, err := config.ParseConfig(configPath)
	if err != nil {
		log.Fatal(err)
	}
	if err := repo.ApplyDatabaseMigrations(config.Database); err != nil {
		log.Fatal(err)
	}
	//go-staticcheck:ignore
	context, cancel := context.WithCancel(context.Background())

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGTERM, syscall.SIGINT)
	pipelineDoneCh, err := initProcessingPipeline(context, *config)
	if err != nil {
		log.Fatal(err)
	}

	select {
	case <-sigChan:
		log.Println("Shutting down")
		cancel()
		<-pipelineDoneCh
	case <-pipelineDoneCh:
		return

	}
}

func initProcessingPipeline(context context.Context, config config.Config) (chan struct{}, error) {
	doneChan := make(chan struct{})
	dbClient, err := repo.NewDbClient(config.Database)
	if err != nil {
		doneChan <- struct{}{}
		return doneChan, err
	}
	repo := repo.NewAudiobookRepository(dbClient)
	pipeline := processing.NewPipeline()
	go pipeline.Start(context, config, doneChan, repo)
	return doneChan, nil

}
