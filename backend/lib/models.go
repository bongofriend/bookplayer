package lib

type Audiobook struct {
	Title       string    `json:"Title"`
	Author      string    `json:"Author"`
	Narrator    string    `json:"Narrator"`
	Description string    `json:"Description"`
	Genre       string    `json:"Genre"`
	Duration    float32   `json:"Duration"`
	Chapters    []Chapter `json:"Chapters"`
	FilePath    string    `json:"FilePath"`
}
type Chapter struct {
	Title     string  `json:"Title"`
	StartTime float32 `json:"StartTime"`
	EndTime   float32 `json:"EndTime"`
	Start     int     `json:"Start"`
	End       int     `json:"End"`
}
