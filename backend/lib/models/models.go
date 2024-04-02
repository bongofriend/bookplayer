package models

type AudiobookCommon struct {
	Title       string  `json:"Title"`
	Author      string  `json:"Author"`
	Narrator    string  `json:"Narrator"`
	Description string  `json:"Description"`
	Genre       string  `json:"Genre"`
	Duration    float32 `json:"Duration"`
}

type ChapterCommon struct {
	Title     string  `json:"Title"`
	StartTime float32 `json:"StartTime"`
	EndTime   float32 `json:"EndTime"`
	Start     int     `json:"Start"`
	End       int     `json:"End"`
}

type Chapter struct {
	ChapterCommon
}

type Audiobook struct {
	AudiobookCommon
	Chapters []Chapter
}

type AudiobookProcessed struct {
	AudiobookCommon
	FilePath          string
	ProcessedChapters []ProcessedChapter
}

type ProcessedChapter struct {
	ChapterCommon
	Numbering int
	FilePath  string
}
