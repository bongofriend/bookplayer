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

	"github.com/bongofriend/bookplayer/backend/lib/config"
	"github.com/bongofriend/bookplayer/backend/lib/models"
	"github.com/bongofriend/bookplayer/backend/lib/processing"
)

type ChapterSplitter struct {
	OutputChan chan processing.AudiobookChapterSplitResult
}

func NewChapterSplitter() (*ChapterSplitter, error) {
	if !ffmpegIsAvailable() {
		return nil, errors.New("ffmpeg is not available")
	}
	return &ChapterSplitter{make(chan processing.AudiobookChapterSplitResult)}, nil
}

func (sp ChapterSplitter) process(config config.ProcessedAudiobooksConfig, input processing.AudiobookMetadataResult) error {
	p := string(input.FilePath)
	audiobook := input.Audiobook
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

	args := getArgs(input, procesedAudiobookPath)
	log.Println(strings.Join(args, " "))
	cmd := exec.Command("ffmpeg", args...)
	if _, err = cmd.Output(); err != nil {
		return err
	}
	chapterPaths := getChapterPaths(audiobook.Chapters, procesedAudiobookPath)
	sp.OutputChan <- processing.AudiobookChapterSplitResult{
		Audiobook:    audiobook,
		DirPath:      procesedAudiobookPath,
		ChapterPaths: chapterPaths,
	}
	return nil
}

func (sp ChapterSplitter) Start(ctx context.Context, wg *sync.WaitGroup, inputChan <-chan processing.AudiobookMetadataResult, config config.ProcessedAudiobooksConfig) {
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

func getArgs(input processing.AudiobookMetadataResult, outputPath string) []string {
	audiobook := input.Audiobook
	filePath := string(input.FilePath)
	endTimes := make([]string, len(audiobook.Chapters))
	for idx, ch := range audiobook.Chapters {
		endTimes[idx] = strconv.FormatFloat(float64(ch.EndTime), 'f', -1, 32)
	}
	endTimeArgs := strings.Join(endTimes, ",")

	outputPathFormat := path.Join(outputPath, "%d.m4b")
	return []string{
		"-i",
		filePath,
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

func getChapterPaths(chapters []models.Chapter, outputPath string) map[string]string {
	panic("unimplemented")
}
