-- +goose Up
CREATE TABLE feeds(
                      id UUID primary key ,
                      created_at timestamp not null,
                      updated_at timestamp not null,
                      name varchar not null unique,
                        url varchar not null unique ,
                    user_id UUID not null references users(id) on delete cascade
);

-- +goose Down
DROP TABLE feeds;