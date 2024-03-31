package chaptersplitter

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path"
	"strconv"
	"strings"
	"sync"

	"github.com/bongofriend/bookplayer/backend/lib"
)

type ChapterSplitter struct {
	OutputChan chan lib.Audiobook
}

func NewChapterSplitter() (*ChapterSplitter, error) {
	if !ffmpegIsAvailable() {
		return nil, errors.New("ffmpeg is not available")
	}
	return &ChapterSplitter{make(chan lib.Audiobook)}, nil
}

func (sp ChapterSplitter) process(config lib.ProcessedAudiobooksConfig, audiobook lib.Audiobook) error {
	p := audiobook.FilePath
	stat, err := os.Stat(p)
	if err != nil {
		return err
	}
	if stat.IsDir() {
		return fmt.Errorf("%s is not file", p)
	}

	stat, err = os.Stat(config.ProcessedPath)
	if err != nil {
		return err
	}
	if !stat.IsDir() {
		return fmt.Errorf("%s already exists as file", config.ProcessedPath)
	}

	procesedAudiobookPath := path.Join(config.ProcessedPath, audiobook.Title)
	if err = os.Mkdir(procesedAudiobookPath, 0755); err != nil {
		return err
	}

	args := getArgs(audiobook, procesedAudiobookPath)
	log.Println(strings.Join(args, " "))
	cmd := exec.Command("ffmpeg", args...)
	if _, err = cmd.Output(); err != nil {
		return err
	}
	sp.OutputChan <- audiobook
	return nil
}

func (sp ChapterSplitter) Start(ctx context.Context, wg *sync.WaitGroup, inputChan <-chan lib.Audiobook, config lib.ProcessedAudiobooksConfig) {
	wg.Add(1)
	go func() {
		defer wg.Done()
		for {
			select {
			case <-ctx.Done():
				close(sp.OutputChan)
				return
			case a := <-inputChan:
				err := sp.process(config, a)
				if err != nil {
					log.Println(err)
				}
			}
		}
	}()
}

func ffmpegIsAvailable() bool {
	_, err := exec.LookPath("ffmpeg")
	return err == nil
}

func getArgs(audiobook lib.Audiobook, outputPath string) []string {
	endTimes := make([]string, len(audiobook.Chapters))
	for idx, ch := range audiobook.Chapters {
		endTimes[idx] = strconv.FormatFloat(float64(ch.EndTime), 'f', -1, 32)
	}
	endTimeArgs := strings.Join(endTimes, ",")

	outputPathFormat := path.Join(outputPath, "chapter%d.m4b")
	return []string{
		"-i",
		audiobook.FilePath,
		"-vn",
		"-acodec",
		"copy",
		"-copyts",
		"-f",
		"segment",
		"-reset_timestamps",
		"1",
		"-segment_start_number",
		"1",
		"-segment_times",
		endTimeArgs,
		outputPathFormat,
	}
}
