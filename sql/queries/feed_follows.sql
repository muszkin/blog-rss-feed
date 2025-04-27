-- name: CreateFeedFollow :one
WITH inserted_feed_follow AS (
    INSERT INTO feed_follows (id, created_at, updated_at, user_id, feed_id)
        values ($1, $2, $3, $4, $5)
        RETURNING *
)
select inserted_feed_follow.*,
       feeds.name as feed_name,
       users.name as user_name
from inserted_feed_follow
inner join feeds on feeds.id = inserted_feed_follow.feed_id
inner join users on users.id = inserted_feed_follow.user_id;

-- name: GetFeedFollowsForUser :many
select feeds.name, users.name
from feed_follows
inner join feeds on feeds.id = feed_follows.feed_id
inner join users on feed_follows.user_id = users.id
where feed_follows.user_id = $1;
