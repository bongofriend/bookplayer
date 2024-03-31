package lib

type Audiobook struct {
	Title       string
	Author      string
	Narrator    string
	Description string
	Genre       string
	Duration    float32
	Chapters    []Chapter

	FilePath string
}

type Chapter struct {
	Title     string
	StartTime float32
	EndTime   float32
	Start     int
	End       int
}
