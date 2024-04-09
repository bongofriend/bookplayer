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

	"github.com/bongofriend/bookplayer/backend/lib/config"
	"github.com/bongofriend/bookplayer/backend/lib/models"
	"github.com/bongofriend/bookplayer/backend/lib/processing"
)

type ChapterSplitter struct {
	resultChan chan models.AudiobookProcessed
	config     config.ProcessedAudiobooksConfig
}

func NewChapterSplitter(config config.ProcessedAudiobooksConfig) (*ChapterSplitter, error) {
	if !ffmpegIsAvailable() {
		return nil, errors.New("ffmpeg is not available")
	}
	return &ChapterSplitter{
		resultChan: make(chan models.AudiobookProcessed),
		config:     config,
	}, nil
}

func (sp ChapterSplitter) process(input processing.AudiobookMetadataResult) error {
	p := input.FilePath
	audiobook := input.Audiobook
	stat, err := os.Stat(p)
	if err != nil {
		return err
	}
	if stat.IsDir() {
		return fmt.Errorf("%s is not file", p)
	}

	stat, err = os.Stat(sp.config.ProcessedPath)
	if err != nil {
		return err
	}
	if !stat.IsDir() {
		return fmt.Errorf("%s already exists as file", sp.config.ProcessedPath)
	}

	procesedAudiobookPath := path.Join(sp.config.ProcessedPath, audiobook.Title)
	if err = os.Mkdir(procesedAudiobookPath, 0755); err != nil {
		return err
	}

	args := getArgs(input, procesedAudiobookPath)
	cmd := exec.Command("ffmpeg", args...)
	if _, err = cmd.Output(); err != nil {
		return err
	}
	processedAudiobook, err := extendAudiobook(audiobook, procesedAudiobookPath, input.FilePath)
	if err != nil {
		return err
	}
	sp.resultChan <- *processedAudiobook
	return nil
}

func ffmpegIsAvailable() bool {
	_, err := exec.LookPath("ffmpeg")
	return err == nil
}

func getChapterOutputPathFormat(dirPath string) string {
	return path.Join(dirPath, "%d.m4b")
}

func getArgs(input processing.AudiobookMetadataResult, outputPath string) []string {
	audiobook := input.Audiobook
	filePath := string(input.FilePath)
	endTimes := make([]string, len(audiobook.Chapters))
	for idx, ch := range audiobook.Chapters {
		endTimes[idx] = strconv.FormatFloat(float64(ch.EndTime), 'f', -1, 32)
	}
	endTimeArgs := strings.Join(endTimes, ",")
	outputPathFormat := getChapterOutputPathFormat(outputPath)

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
		"0",
		"-segment_times",
		endTimeArgs,
		outputPathFormat,
	}
}

func extendAudiobook(a models.Audiobook, splitChapterDirPath string, audiobookFilePath string) (*models.AudiobookProcessed, error) {
	processedChapters := make([]models.ProcessedChapter, 0)
	outputPathFormat := getChapterOutputPathFormat(splitChapterDirPath)
	for _, ch := range a.Chapters {
		chapterPath := fmt.Sprintf(outputPathFormat, ch.Numbering)
		stat, err := os.Stat(chapterPath)
		if err != nil {
			return nil, err
		}
		if !stat.Mode().IsRegular() {
			return nil, fmt.Errorf("%s not a file", chapterPath)
		}
		processed := models.ProcessedChapter{
			ChapterCommon: ch.ChapterCommon,
			FilePath:      chapterPath,
		}
		processedChapters = append(processedChapters, processed)

	}
	return &models.AudiobookProcessed{
		AudiobookCommon:   a.AudiobookCommon,
		FilePath:          audiobookFilePath,
		ProcessedChapters: processedChapters,
	}, nil
}

func (sp ChapterSplitter) Shutdown() {
	close(sp.resultChan)
}

func (sp ChapterSplitter) Output() (chan models.AudiobookProcessed, error) {
	return sp.resultChan, nil
}

func (sp ChapterSplitter) Start(ctx context.Context, inputChan <-chan processing.AudiobookMetadataResult, doneCh chan struct{}) {
	go func() {
		defer func() {
			sp.Shutdown()
			doneCh <- struct{}{}
		}()
		for {
			select {
			case <-ctx.Done():
				return
			case a := <-inputChan:
				err := sp.process(a)
				if err != nil {
					log.Println(err)
				}
			}
		}
	}()
}
