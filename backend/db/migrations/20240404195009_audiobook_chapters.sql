-- +goose Up
-- +goose StatementBegin
Create Table Audiobook (
    id integer primary key not null,
    title text not null,
    author text not null,
    narrator text not null,
    description text not null,
    duration int not null,
    dir_path text not null,
    chapter_count int not null,
    genre text not null
);

Create Table Chapter (
    id integer primary key not null,
    audiobook_id int not null,
    numbering int not null,
    title text not null,
    start_time float not null,
    end_time float not null,
    file_path text not null,

    foreign key(audiobook_id) references Audiobooks(id)
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
Drop Table Chapter;
Drop Table Audiobook;
-- +goose StatementEnd
