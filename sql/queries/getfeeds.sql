-- name: GetFeeds :many
SELECT feeds.id, feeds.created_at, feeds.updated_at, feeds.name, url, users.name as user_name
FROM feeds
LEFT JOIN users ON feeds.user_id=users.id;
