package metadataextractor

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"os"
	"os/exec"
	"strconv"

	"github.com/bongofriend/bookplayer/backend/lib/models"
	"github.com/bongofriend/bookplayer/backend/lib/processing"
)

type MetadataExtractor struct {
	metadataChan chan processing.AudiobookMetadataResult
}

type Chapter struct {
	ID        int    `json:"id"`
	TimeBase  string `json:"time_base"`
	Start     int    `json:"start"`
	StartTime string `json:"start_time"`
	End       int    `json:"end"`
	EndTime   string `json:"end_time"`
	Tags      struct {
		Title string `json:"title"`
	} `json:"tags"`
}

type Format struct {
	Filename       string `json:"filename"`
	NbStreams      int    `json:"nb_streams"`
	NbPrograms     int    `json:"nb_programs"`
	FormatName     string `json:"format_name"`
	FormatLongName string `json:"format_long_name"`
	StartTime      string `json:"start_time"`
	Duration       string `json:"duration"`
	Size           string `json:"size"`
	BitRate        string `json:"bit_rate"`
	ProbeScore     int    `json:"probe_score"`
	Tags           Tags   `json:"tags"`
}

type Tags struct {
	MajorBrand       string `json:"major_brand"`
	MinorVersion     string `json:"minor_version"`
	CompatibleBrands string `json:"compatible_brands"`
	Title            string `json:"title"`
	Artist           string `json:"artist"`
	Composer         string `json:"composer"`
	Album            string `json:"album"`
	Encoder          string `json:"encoder"`
	Comment          string `json:"comment"`
	Genre            string `json:"genre"`
	MediaType        string `json:"media_type"`
}

type AudiobookMetadata struct {
	Chapters []Chapter `json:"chapters"`
	Format   Format    `json:"format"`
}

func NewMetadataExtractor() (*MetadataExtractor, error) {
	if !ffprobeIsAvailable() {
		return nil, errors.New("ffmpeg is not installed or found")
	}
	return &MetadataExtractor{
		metadataChan: make(chan processing.AudiobookMetadataResult),
	}, nil
}

func (m MetadataExtractor) extractMetadata(path string) error {
	filePath := string(path)
	if stat, err := os.Stat(string(filePath)); err != nil || stat.IsDir() {
		if err != nil {
			log.Println(err)
		}
	}
	ffprobeArgs := []string{"-print_format", "json", "-show_format", "-show_chapters", filePath}
	cmd := exec.Command("ffprobe", ffprobeArgs...)
	output, err := cmd.Output()
	if err != nil {
		return err
	}
	ffprobeOutput := AudiobookMetadata{}
	if err := json.Unmarshal(output, &ffprobeOutput); err != nil {
		return err
	}
	model, err := ffprobeOutput.AsModel()
	if err != nil {
		return err
	}
	m.metadataChan <- processing.AudiobookMetadataResult{
		Audiobook: model,
		FilePath:  path,
	}
	return nil
}

func ffprobeIsAvailable() bool {
	_, err := exec.LookPath("ffmpeg")
	return err == nil
}

func (a AudiobookMetadata) AsModel() (models.Audiobook, error) {
	tags := a.Format.Tags
	audiobookDuration, err := strconv.ParseFloat(a.Format.Duration, 32)
	if err != nil {
		return models.Audiobook{}, err
	}

	chapters := make([]models.Chapter, len(a.Chapters))
	for idx, c := range a.Chapters {
		data, err := c.asModel()
		if err != nil {
			return models.Audiobook{}, nil
		}
		chapters[idx] = data
	}

	return models.Audiobook{
		AudiobookCommon: models.AudiobookCommon{
			Title:       tags.Title,
			Author:      tags.Artist,
			Narrator:    tags.Composer,
			Description: tags.Comment,
			Genre:       tags.Genre,
			Duration:    float32(audiobookDuration),
		},
		Chapters: chapters,
	}, nil
}

func (c Chapter) asModel() (models.Chapter, error) {
	startTime, err := strconv.ParseFloat(c.StartTime, 32)
	if err != nil {
		return models.Chapter{}, err
	}
	endTime, err := strconv.ParseFloat(c.EndTime, 32)
	if err != nil {
		return models.Chapter{}, err
	}

	return models.Chapter{
		ChapterCommon: models.ChapterCommon{
			Title:     c.Tags.Title,
			StartTime: float32(startTime),
			EndTime:   float32(endTime),
			Start:     c.Start,
			End:       c.End,
			Numbering: c.ID,
		},
	}, nil
}

func (m MetadataExtractor) Shutdown() {
	log.Println("Shutting down MetadataExtractor")
	close(m.metadataChan)
}

func (m MetadataExtractor) Output() (chan processing.AudiobookMetadataResult, error) {
	return m.metadataChan, nil
}

func (m *MetadataExtractor) Start(ctx context.Context, intputChan chan string, doneChan chan struct{}) {
	go func() {
		for {
			select {
			case <-ctx.Done():
				m.Shutdown()
				doneChan <- struct{}{}
				return
			case path, ok := <-intputChan:
				if !ok {
					return
				}
				if err := m.extractMetadata(path); err != nil {
					log.Println(err)
				}

			}
		}
	}()
}
