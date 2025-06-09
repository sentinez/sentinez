-- name: Get :one
SELECT * FROM users
WHERE id = $1 LIMIT 1;

-- name: GetMany :many
SELECT * FROM users
ORDER BY username;

-- name: Delete :exec
DELETE FROM users
WHERE id = $1;