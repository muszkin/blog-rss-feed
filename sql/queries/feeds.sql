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
-- name: MarkFeedFetched :exec
update feeds
set last_fetched_at = now() and updated_at = now()
where id = $1;
-- name: GetNextFeedToFetch :one
select * from feeds order by last_fetched_at nulls first limit 1;