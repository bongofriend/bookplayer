package processing

import "github.com/bongofriend/bookplayer/backend/lib/models"

type AudiobookMetadataResult struct {
	Audiobook models.Audiobook
	FilePath  string
}

type AudiobookChapterSplitResult struct {
	Audiobook    models.AudiobookProcessed
	DirPath      string
	ChapterPaths map[string]string
}
