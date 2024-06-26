
-- name: InsertAudiobook :execresult
Insert Into Audiobook (title, author, narrator, description, duration, dir_path, chapter_count, genre) Values (?, ?, ?, ?, ?, ?, ?, ?);

-- name: InsertChapter :exec
Insert Into Chapter (audiobook_id, title, numbering, start_time, end_time, file_path) Values (?, ?, ?, ?, ?, ?);

-- name: GetAllAudiobooks :many
Select *
From Audiobook a;

-- name: GetAudiobookChapters :many
Select *
From Chapter c
Where c.audiobook_id = ?
Order By numbering Asc;

-- name: GetAudiobookById :many
Select sqlc.embed(a), sqlc.embed(c)
From Audiobook a
Join Chapter c On a.id = c.audiobook_id
Where a.id = ?