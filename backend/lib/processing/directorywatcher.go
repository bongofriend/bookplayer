package processing

import (
	"bytes"
	"crypto/sha1"
	"encoding/gob"
	"encoding/hex"
	"io"
	"log"
	"os"
	"path"
	"path/filepath"

	"github.com/bongofriend/bookplayer/backend/lib/config"
)

const saveFileName = "seen_files"

type DirectoryWatcher struct {
	config     config.Config
	fileHashes map[string]string
}

func NewDirectoryWatcher(c config.Config) (*DirectoryWatcher, error) {
	if err := os.MkdirAll(c.AudiobookDirectory, 0777); err != nil {
		return nil, err
	}
	audiobooks, err := loadSeenAudiobooks(c)
	if err != nil {
		return nil, err
	}
	return &DirectoryWatcher{
		fileHashes: audiobooks,
		config:     c,
	}, nil
}

func loadSeenAudiobooks(c config.Config) (map[string]string, error) {
	saveFilePath := path.Join(c.ApplicationDirectory, saveFileName)
	_, err := os.Stat(saveFilePath)
	if err != nil {
		if os.IsNotExist(err) {
			return make(map[string]string), nil
		}
		return nil, err
	}
	data, err := os.ReadFile(saveFilePath)
	if err != nil {
		return nil, err
	}
	var seenAudiobooks map[string]string
	dec := gob.NewDecoder(bytes.NewReader(data))
	if err := dec.Decode(&seenAudiobooks); err != nil {
		return nil, err
	}
	return seenAudiobooks, nil
}

func saveSeenAudiobooks(c config.Config, audiobookFiles map[string]string) error {
	buf := bytes.Buffer{}
	enc := gob.NewEncoder(&buf)
	if err := enc.Encode(audiobookFiles); err != nil {
		return err
	}
	saveFilePath := path.Join(c.ApplicationDirectory, saveFileName)
	err := os.WriteFile(saveFilePath, buf.Bytes(), 0777)
	return err
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
		if ext := filepath.Ext(name); ext != ".m4b" {
			continue
		}
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
	log.Println("Shutting down DirectoryWatcher")
	saveSeenAudiobooks(d.config, d.fileHashes)
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
