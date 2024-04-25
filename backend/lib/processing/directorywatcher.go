package processing

import (
	"crypto/sha1"
	"encoding/hex"
	"io"
	"log"
	"os"
	"path/filepath"

	"github.com/bongofriend/bookplayer/backend/lib/config"
)

type DirectoryWatcher struct {
	config     config.Config
	fileHashes map[string]string
}

func NewDirectoryWatcher(c config.Config) DirectoryWatcher {
	return DirectoryWatcher{
		fileHashes: make(map[string]string),
		config:     c,
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

func (d *DirectoryWatcher) ProcessInput(input struct{}, outputChan chan string) error {
	paths, err := os.ReadDir(d.config.AudiobookDirectory)
	if err != nil {
		return err
	}
	for _, p := range paths {
		if p.IsDir() {
			continue
		}
		name := p.Name()
		pathToFile := filepath.Join(d.config.AudiobookDirectory, name)
		fileHash, found := d.fileHashes[name]
		hash, err := fileCheckSum(pathToFile)
		if err != nil {
			return err
		}
		if !found || hash != fileHash {
			d.fileHashes[name] = hash
			outputChan <- pathToFile
		}
	}
	return nil
}

func (d DirectoryWatcher) Shutdown() {
	log.Printf("Stopping to watch %s", d.config.AudiobookDirectory)
}

func (d DirectoryWatcher) CommandsToReceive() []PipelineCommandType {
	return []PipelineCommandType{
		Scan,
	}
}

func (d DirectoryWatcher) ProcessCommand(cmd PipelineCommand, inputChan chan struct{}, outputChan chan string) error {
	if cmd.CmdType != Scan {
		return nil
	}
	return d.ProcessInput(struct{}{}, outputChan)
}
