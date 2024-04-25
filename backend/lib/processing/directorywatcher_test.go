package processing_test

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/bongofriend/bookplayer/backend/lib/config"
	"github.com/bongofriend/bookplayer/backend/lib/processing"
)

func TestDirectoryWatcherObserve(t *testing.T) {
	context, cancel := context.WithCancel(context.Background())
	testConfig := config.Config{
		AudiobookDirectory: t.TempDir(),
		ScanInterval:       2 * time.Second,
	}
	handler := processing.NewDirectoryWatcher(testConfig)
	watcher := processing.NewPipelineStage[struct{}, string](&handler)
	doneConsumer := make(chan struct{})
	ticker := time.NewTicker(testConfig.ScanInterval)

	go func() {
		defer ticker.Stop()
		for {
			select {
			case <-context.Done():
				return
			case <-ticker.C:
				watcher.InputChan <- struct{}{}
			}
		}
	}()

	go watcher.Start(context)

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
	<-watcher.DoneChan

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
	handler := processing.NewDirectoryWatcher(testConfig)
	watcher := processing.NewPipelineStage[struct{}, string](&handler)
	doneCh := make(chan struct{})
	ticker := time.NewTicker(testConfig.ScanInterval)

	go func() {
		defer ticker.Stop()
		for {
			select {
			case <-context.Done():
				return
			case <-ticker.C:
				watcher.InputChan <- struct{}{}
			}
		}
	}()

	go watcher.Start(context)

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
	<-doneCh
	<-watcher.DoneChan

	if filePathReceivedCount != 2 {
		t.Fatalf("Expected %d emissions; received: %d", 2, filePathReceivedCount)
	}
}
