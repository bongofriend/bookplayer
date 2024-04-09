package directorywatcher

import (
	"context"
	"crypto/sha1"
	"encoding/hex"
	"io"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/bongofriend/bookplayer/backend/lib/config"
)

type DirectoryWatcher struct {
	config     config.AudiobooksConfig
	pathChan   chan string
	fileHashes map[string]string
}

func NewDirectoryWatcher(c config.AudiobooksConfig) DirectoryWatcher {
	return DirectoryWatcher{
		pathChan:   make(chan string),
		fileHashes: make(map[string]string),
		config:     c,
	}
}

// TODO Load processed file hashes on initialization
func (d DirectoryWatcher) load() {

}

// TODO Dump file hashes on shutdown
func (d DirectoryWatcher) flush() {

}

func (d DirectoryWatcher) parseDirectoryContent(c config.AudiobooksConfig) error {
	paths, err := os.ReadDir(c.AudibookDirectoryPath)
	if err != nil {
		return err
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
			return err
		}
		if !found || hash != fileHash {
			d.fileHashes[name] = hash
			d.pathChan <- pathToFile
		}
	}
	return nil
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

func (d DirectoryWatcher) Shutdown() {
	log.Printf("Stopping to watch %s", d.config.AudibookDirectoryPath)
	close(d.pathChan)
	d.flush()
}

func (d DirectoryWatcher) Output() (chan string, error) {
	return d.pathChan, nil
}

func (d *DirectoryWatcher) Start(ctx context.Context, inputChan chan struct{}, doneCh chan struct{}) {
	c := d.config
	ticker := time.NewTicker(c.Interval)
	d.load()
	log.Printf("Watching %s", c.AudibookDirectoryPath)
	go func() {
		defer func() {
			d.Shutdown()
			doneCh <- struct{}{}
		}()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				if err := d.parseDirectoryContent(c); err != nil {
					log.Println(err)
				}
			}
		}
	}()
}
