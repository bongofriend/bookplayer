Create Table Audiobook (
    id int primary key,
    title text not null,
    author text not null,
    narrator text not null,
    description text not null,
    duration int not null,
    dir_path text not null,
    chapter_count int not null
);

Create Table Chapter (
    id int primary key,
    audiobook_id int not null,
    numbering int not null,
    title text not null,
    start_time float not null,
    end_time float not null,
    file_path text not null,

    foreign key(audiobook_id) references Audiobooks(id)
);