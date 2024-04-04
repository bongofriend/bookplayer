package repo

import (
	"context"

	"github.com/bongofriend/bookplayer/backend/lib/data/datasource"
	"github.com/bongofriend/bookplayer/backend/lib/models"
)

type AudiobookRepository struct {
	client *DbClient
}

func NewAudiobookRepository(client *DbClient) *AudiobookRepository {
	return &AudiobookRepository{client}
}

func (r *AudiobookRepository) InsertAudiobook(context context.Context, audiobook models.AudiobookProcessed) error {
	tx, err := r.client.db.Begin()
	if err != nil {
		return nil
	}
	defer tx.Commit()
	qtx := r.client.queries.WithTx(tx)
	audiobookParams := audiobookAsParams(audiobook)
	res, err := qtx.InsertAudiobook(context, audiobookParams)
	if err != nil {
		return err
	}
	id, err := res.LastInsertId()
	if err != nil {
		return err
	}
	for _, params := range chaptersAsParams(id, audiobook) {
		if err := qtx.InsertChapter(context, params); err != nil {
			return err
		}
	}
	return nil
}

func audiobookAsParams(audiobook models.AudiobookProcessed) datasource.InsertAudiobookParams {
	return datasource.InsertAudiobookParams{
		Title:        audiobook.Title,
		Author:       audiobook.Author,
		Narrator:     audiobook.Narrator,
		Description:  audiobook.Description,
		Duration:     int64(audiobook.Duration),
		DirPath:      audiobook.FilePath,
		ChapterCount: int64(len(audiobook.ProcessedChapters)),
	}

}

func chaptersAsParams(id int64, audiobook models.AudiobookProcessed) []datasource.InsertChapterParams {
	params := make([]datasource.InsertChapterParams, len(audiobook.ProcessedChapters))
	for _, ch := range audiobook.ProcessedChapters {
		p := datasource.InsertChapterParams{
			AudiobookID: id,
			Title:       ch.Title,
			Numbering:   int64(ch.Numbering),
			StartTime:   float64(ch.StartTime),
			EndTime:     float64(ch.EndTime),
			FilePath:    ch.FilePath,
		}
		params = append(params, p)
	}
	return params
}
