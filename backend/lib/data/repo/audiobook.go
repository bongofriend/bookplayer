package repo

import (
	"context"
	"fmt"
	"slices"

	"github.com/bongofriend/bookplayer/backend/lib/data/datasource"
	"github.com/bongofriend/bookplayer/backend/lib/models"
)

type AudiobookRepository struct {
	client *DbClient
}

func NewAudiobookRepository(client *DbClient) *AudiobookRepository {
	return &AudiobookRepository{client}
}

func (r *AudiobookRepository) InsertAudiobook(context context.Context, audiobook models.AudiobookProcessed) (int64, error) {
	tx, err := r.client.db.Begin()
	if err != nil {
		return -1, nil
	}
	qtx := r.client.queries.WithTx(tx)
	audiobookParams := audiobookAsParams(audiobook)
	res, err := qtx.InsertAudiobook(context, audiobookParams)
	if err != nil {
		tx.Rollback()
		return -1, err
	}
	id, err := res.LastInsertId()
	if err != nil {
		tx.Rollback()
		return -1, err
	}
	for _, params := range chaptersAsParams(id, audiobook) {
		if err := qtx.InsertChapter(context, params); err != nil {
			tx.Rollback()
			return -1, err
		}
	}
	tx.Commit()
	return id, nil
}

func (r *AudiobookRepository) GetAudiobookById(context context.Context, id int64) (*models.AudiobookProcessed, error) {
	rows, err := r.client.queries.GetAudiobookById(context, id)
	if err != nil {
		return nil, err
	}
	if len(rows) == 0 {
		return nil, fmt.Errorf("audiobook with id %d not found", id)
	}
	model := audiobookRowsToModels(rows)
	return &model, nil
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
		Genre:        audiobook.Genre,
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

func audiobookRowsToModels(rows []datasource.GetAudiobookByIdRow) models.AudiobookProcessed {
	chapters := make([]models.ProcessedChapter, len(rows))
	for idx, r := range rows {
		chapter := models.ProcessedChapter{
			ChapterCommon: models.ChapterCommon{
				Title:     r.Chapter.Title,
				StartTime: float32(r.Chapter.StartTime),
				EndTime:   float32(r.Chapter.EndTime),
				//Start:     0,
				//End:       0,
				Numbering: int(r.Chapter.Numbering),
			},
			FilePath: r.Chapter.FilePath,
		}
		chapters[idx] = chapter
	}
	slices.SortFunc(chapters, func(c1 models.ProcessedChapter, c2 models.ProcessedChapter) int {
		return c1.Numbering - c2.Numbering
	})
	firstRow := rows[0]
	return models.AudiobookProcessed{
		AudiobookCommon: models.AudiobookCommon{
			Title:       firstRow.Audiobook.Title,
			Author:      firstRow.Audiobook.Author,
			Narrator:    firstRow.Audiobook.Narrator,
			Description: firstRow.Audiobook.Description,
			Genre:       firstRow.Audiobook.Genre,
			Duration:    float32(firstRow.Audiobook.Duration),
		},
		FilePath:          firstRow.Chapter.FilePath,
		ProcessedChapters: chapters,
	}
}
