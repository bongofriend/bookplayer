package directorywatcher_test

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/bongofriend/bookplayer/backend/lib/config"
	"github.com/bongofriend/bookplayer/backend/lib/processing"
	"github.com/bongofriend/bookplayer/backend/lib/processing/directorywatcher"
)

func TestDirectoryWatcherObserve(t *testing.T) {
	context, cancel := context.WithCancel(context.Background())
	testConfig := config.Config{
		AudiobookDirectory: t.TempDir(),
		ScanInterval:       2 * time.Second,
	}
	handler := directorywatcher.NewDirectoryWatcher(testConfig)
	watcher := processing.NewPipelineComponent[time.Time, string](&handler)
	doneConsumer := make(chan struct{})
	doneWatcher := make(chan struct{})
	ticker := time.NewTicker(testConfig.ScanInterval)
	watcher.Start(context, ticker.C, doneWatcher)

	testFileName := "test.txt"
	testFileContent := "Hello from TestFile!"
	testFilePath := filepath.Join(testConfig.AudiobookDirectory, testFileName)
	expectedFilePathReceived := false

	go func() {
		defer func() {
			doneConsumer <- struct{}{}
		}()
		output := watcher.OutputChan
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
	testConfig := config.Config{
		AudiobookDirectory: t.TempDir(),
		ScanInterval:       2 * time.Second,
	}
	handler := directorywatcher.NewDirectoryWatcher(testConfig)
	watcher := processing.NewPipelineComponent[time.Time, string](&handler)
	doneCh := make(chan struct{})
	ticker := time.NewTicker(testConfig.ScanInterval)
	watcher.Start(context, ticker.C, doneCh)

	testFileName := "test.txt"
	testFilePath := filepath.Join(testConfig.AudiobookDirectory, testFileName)
	filePathReceivedCount := 0

	go func() {
		defer func() {
			doneCh <- struct{}{}
		}()
		output := watcher.OutputChan
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
		time.Sleep(2 * testConfig.ScanInterval)
	}

	cancel()
	for i := 0; i < 2; i++ {
		<-doneCh
	}

	if filePathReceivedCount != 2 {
		t.Fatalf("Expected %d emissions; received: %d", 2, filePathReceivedCount)
	}
}
