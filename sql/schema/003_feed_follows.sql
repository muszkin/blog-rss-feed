-- +goose Up
CREATE TABLE feed_follows(
                      id UUID primary key ,
                      created_at timestamp not null,
                      updated_at timestamp not null,
                      user_id UUID not null references users(id) on delete cascade,
                      feed_id uuid not null references feeds(id) on delete cascade,
                      unique(user_id, feed_id)
);
alter table feeds
    drop constraint feeds_user_id_fkey;
alter table feeds
    drop column user_id;
alter table feeds
    drop constraint feeds_name_key;


-- +goose Down
DROP TABLE feed_follows;
alter table feeds
    add column user_id uuid references users(id) on delete cascade;
