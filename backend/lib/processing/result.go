package processing

import "github.com/bongofriend/bookplayer/backend/lib/models"

type AudiobookDiscoveryResult string

type AudiobookMetadataResult struct {
	Audiobook models.Audiobook
	FilePath  AudiobookDiscoveryResult
}

type AudiobookChapterSplitResult struct {
	Audiobook    models.Audiobook
	DirPath      string
	ChapterPaths map[string]string
}
