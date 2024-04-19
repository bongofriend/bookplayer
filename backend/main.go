package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/bongofriend/bookplayer/backend/lib/config"
	"github.com/bongofriend/bookplayer/backend/lib/data/repo"
	_ "github.com/mattn/go-sqlite3"
)

func main() {
	envPath, err := config.GetEnvPathFromFlags()
	if err != nil {
		log.Fatal(err)
	}

	config, err := config.ParseConfig(envPath)
	if err != nil {
		log.Fatal(err)
	}
	if err := repo.ApplyDatabaseMigrations(config.Database); err != nil {
		log.Fatal(err)
	}
	//go-staticcheck:ignore
	_, cancel := context.WithCancel(context.Background())

	wg := sync.WaitGroup{}
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGTERM, syscall.SIGINT)

	<-sigChan
	log.Println("Shutting down")
	cancel()
	wg.Wait()
}
