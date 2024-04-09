package directorywatcher_test

import (
	"context"
	"log"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/bongofriend/bookplayer/backend/lib/config"
	"github.com/bongofriend/bookplayer/backend/lib/processing/directorywatcher"
)

func TestDirectoryWatcherObserve(t *testing.T) {
	context, cancel := context.WithCancel(context.Background())
	testConfig := config.AudiobooksConfig{
		AudibookDirectoryPath: t.TempDir(),
		Interval:              2 * time.Second,
	}
	watcher := directorywatcher.NewDirectoryWatcher(testConfig)
	doneConsumer := make(chan struct{})
	doneWatcher := make(chan struct{})
	watcher.Start(context, make(chan struct{}), doneWatcher)

	testFileName := "test.txt"
	testFileContent := "Hello from TestFile!"
	testFilePath := filepath.Join(testConfig.AudibookDirectoryPath, testFileName)
	expectedFilePathReceived := false

	go func() {
		defer func() {
			doneConsumer <- struct{}{}
		}()
		output, err := watcher.Output()
		if err != nil {
			log.Fatal(err)
		}
		select {
		case <-context.Done():
			return
		case p := <-output:
			if p == testFilePath {
				expectedFilePathReceived = true
			}
		}

	}()

	os.WriteFile(testFilePath, []byte(testFileContent), 0666)
	<-doneConsumer
	cancel()
	<-doneWatcher

	if !expectedFilePathReceived {
		t.Fatal("Expected test file path not received")
	}
}

func TestDirectoryWatcherUniqueFiles(t *testing.T) {
	context, cancel := context.WithCancel(context.Background())
	testConfig := config.AudiobooksConfig{
		AudibookDirectoryPath: t.TempDir(),
		Interval:              2 * time.Second,
	}
	watcher := directorywatcher.NewDirectoryWatcher(testConfig)
	doneCh := make(chan struct{})
	watcher.Start(context, make(chan struct{}), doneCh)

	testFileName := "test.txt"
	testFilePath := filepath.Join(testConfig.AudibookDirectoryPath, testFileName)
	filePathReceivedCount := 0

	go func() {
		defer func() {
			doneCh <- struct{}{}
		}()
		output, err := watcher.Output()
		if err != nil {
			log.Fatal(err)
		}
		for {
			select {
			case <-context.Done():
				return
			case p := <-output:
				if p == testFilePath {
					filePathReceivedCount += 1
				}
			}
		}
	}()

	textToWrite := []string{"Hello", "Hello", "World"}
	for _, text := range textToWrite {
		os.WriteFile(testFilePath, []byte(text), 0644)
		time.Sleep(2 * testConfig.Interval)
	}

	cancel()
	for i := 0; i < 2; i++ {
		<-doneCh
	}

	if filePathReceivedCount != 2 {
		t.Fatalf("Expected %d emissions; received: %d", 2, filePathReceivedCount)
	}
}
