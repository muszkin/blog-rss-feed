-- +goose Up
create table posts(
    id uuid not null primary key,
    created_at timestamp not null,
    updated_at timestamp not null,
    title text not null,
    url text not null unique,
    description text null,
    published_at timestamp not null,
    feed_id uuid not null references feeds(id) on delete cascade
);

-- +goose Down
drop table posts;