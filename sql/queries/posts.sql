-- name: CreatePost :one
INSERT INTO posts (id, created_at, updated_at, title, url, description, published_at, feed_id)
VALUES (
           $1,
           $2,
           $3,
           $4,
           $5,
            $6,
            $7,
        $8
       )
on conflict (url) do update
    set updated_at = now(),
        title = $4,
        description = $6,
        published_at = $7
RETURNING *;
-- name: GetPostsForUser :many
select posts.* from posts
inner join feed_follows on feed_follows.feed_id = posts.feed_id
inner join users on feed_follows.user_id = users.id
where users.id = $1
order by posts.published_at desc
limit $2;
