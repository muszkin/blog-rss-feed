-- name: CreateFeed :one
INSERT INTO feeds (id, created_at, updated_at, name, url)
VALUES (
           $1,
           $2,
           $3,
           $4,
            $5
       )
RETURNING *;
-- name: GetFeedByName :one
SELECT * FROM feeds WHERE name = $1;
-- name: GetFeedById :one
SELECT * FROM feeds WHERE id = $1;
-- name: GetFeedByUrl :one
SELECT * FROM feeds WHERE url = $1;
-- name: GetFeeds :many
SELECT * FROM feeds;