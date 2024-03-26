package metadataextractor

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"os"
	"os/exec"
	"sync"
)

type MetadataExtractor struct {
	MetadataChan chan AudiobookMetadata
}

type AudiobookMetadata struct {
	Chapters []struct {
		ID        int    `json:"id"`
		TimeBase  string `json:"time_base"`
		Start     int    `json:"start"`
		StartTime string `json:"start_time"`
		End       int    `json:"end"`
		EndTime   string `json:"end_time"`
		Tags      struct {
			Title string `json:"title"`
		} `json:"tags"`
	} `json:"chapters"`
	Format struct {
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
		Tags           struct {
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
		} `json:"tags"`
	} `json:"format"`
}

func NewMetadataExtractor() (MetadataExtractor, error) {
	if !ffprobeIsAvailable() {
		return MetadataExtractor{}, errors.New("ffmpeg is not installed or found")
	}
	return MetadataExtractor{
		MetadataChan: make(chan AudiobookMetadata),
	}, nil
}

// TODO
func (m MetadataExtractor) extractMetadata(path string) {
	if stat, err := os.Stat(path); err != nil || stat.IsDir() {
		if err != nil {
			log.Println(err)
		}
	}
	ffprobeArgs := []string{"-print_format", "json", "-show_format", "-show_chapters", path}
	cmd := exec.Command("ffprobe", ffprobeArgs...)
	output, err := cmd.Output()
	if err != nil {
		log.Println(err)
		return
	}
	ffprobeOutput := AudiobookMetadata{}
	if err := json.Unmarshal(output, &ffprobeOutput); err != nil {
		log.Println(err)
		return
	}
	m.MetadataChan <- ffprobeOutput
}

func (m MetadataExtractor) Process(ctx context.Context, wg *sync.WaitGroup, pathChan <-chan string) {
	wg.Add(1)
	go func() {
		defer wg.Done()
		for {
			select {
			case <-ctx.Done():
				log.Println("Shutting down MetadataExtractor")
				return
			case path := <-pathChan:
				m.extractMetadata(path)
			}
		}
	}()
}

func ffprobeIsAvailable() bool {
	_, err := exec.LookPath("ffmpeg")
	return err == nil
}
