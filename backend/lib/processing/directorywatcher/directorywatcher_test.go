package directorywatcher_test

import (
	"context"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"github.com/bongofriend/bookplayer/backend/lib/config"
	"github.com/bongofriend/bookplayer/backend/lib/processing"
	"github.com/bongofriend/bookplayer/backend/lib/processing/directorywatcher"
)

func TestDirectoryWatcherObserve(t *testing.T) {
	watcher := directorywatcher.NewDirectoryWatcher()
	context, cancel := context.WithCancel(context.Background())
	testConfig := config.AudiobooksConfig{
		AudibookDirectoryPath: t.TempDir(),
		Interval:              2 * time.Second,
	}
	var wg *sync.WaitGroup = &sync.WaitGroup{}

	wg.Add(1)
	err := watcher.Start(context, wg, testConfig)
	if err != nil {
		wg.Done()
		t.Fatal(err)
	}
	testFileName := "test.txt"
	testFileContent := "Hello from TestFile!"
	testFilePath := filepath.Join(testConfig.AudibookDirectoryPath, testFileName)
	expectedFilePathReceived := false

	go func() {
		defer wg.Done()
		select {
		case <-context.Done():
			return
		case p := <-watcher.PathChan:
			if p == processing.AudiobookDiscoveryResult(testFilePath) {
				expectedFilePathReceived = true
			}
		}

	}()

	os.WriteFile(testFilePath, []byte(testFileContent), 0666)
	time.Sleep(testConfig.Interval * 2)
	cancel()
	wg.Wait()
	if !expectedFilePathReceived {
		t.Fatal("Expected test file path not received")
	}
}

func TestDirectoryWatcherUniqueFiles(t *testing.T) {
	watcher := directorywatcher.NewDirectoryWatcher()
	context, cancel := context.WithCancel(context.Background())
	testConfig := config.AudiobooksConfig{
		AudibookDirectoryPath: t.TempDir(),
		Interval:              2 * time.Second,
	}
	var wg *sync.WaitGroup = &sync.WaitGroup{}

	err := watcher.Start(context, wg, testConfig)
	if err != nil {
		wg.Done()
		t.Fatal(err)
	}

	testFileName := "test.txt"
	testFilePath := filepath.Join(testConfig.AudibookDirectoryPath, testFileName)
	filePathReceivedCount := 0

	wg.Add(1)
	go func() {
		defer wg.Done()
		for {
			select {
			case <-context.Done():
				return
			case p := <-watcher.PathChan:
				if p == processing.AudiobookDiscoveryResult(testFilePath) {
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
	wg.Wait()

	if filePathReceivedCount != 2 {
		t.Fatal("Received unexpected number of files")
	}
}
