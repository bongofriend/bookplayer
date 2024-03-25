package directorywatcher

import (
	"context"
	"crypto/sha1"
	"encoding/hex"
	"io"
	"log"
	"os"
	"path/filepath"
	"sync"
	"time"

	config "github.com/bongofriend/backend/lib"
)

type DirectoryWatcher struct {
	PathChan   chan string
	fileHashes map[string]string
}

func NewDirectoryWatcher() DirectoryWatcher {
	return DirectoryWatcher{
		PathChan:   make(chan string),
		fileHashes: make(map[string]string),
	}
}

// TODO Load processed file hashes on initialization
func (d DirectoryWatcher) load() {

}

// TODO Dump file hashes on shutdown
func (d DirectoryWatcher) flush() {

}

func (d DirectoryWatcher) shutdown(c config.AudiobooksConfig) {
	log.Printf("Stopping to watch %s", c.AudibookDirectoryPath)
	d.flush()
}

func (d DirectoryWatcher) parseDirectoryContent(c config.AudiobooksConfig) {
	paths, err := os.ReadDir(c.AudibookDirectoryPath)
	if err != nil {
		log.Println(err)
		return
	}
	for _, p := range paths {
		if p.IsDir() {
			continue
		}
		name := p.Name()
		pathToFile := filepath.Join(c.AudibookDirectoryPath, name)
		fileHash, found := d.fileHashes[name]
		hash, err := fileCheckSum(pathToFile)
		if err != nil {
			log.Println(err)
			continue
		}
		if !found || hash != fileHash {
			d.fileHashes[name] = hash
			d.PathChan <- pathToFile
		}
	}
}

func fileCheckSum(p string) (string, error) {
	file, err := os.Open(p)
	if err != nil {
		return "", err
	}
	defer file.Close()

	buf := make([]byte, 1024*1024)
	hash := sha1.New()
	if _, err := io.CopyBuffer(hash, file, buf); err != nil {
		return "", err
	}
	return hex.EncodeToString(hash.Sum(nil)), nil
}

func (d DirectoryWatcher) Start(ctx context.Context, wg *sync.WaitGroup, c config.AudiobooksConfig) error {
	d.load()
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
				d.parseDirectoryContent(c)
			}
		}
	}()
	return nil
}
