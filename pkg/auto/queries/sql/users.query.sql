-- name: Get :one
SELECT * FROM users
WHERE id = $1 LIMIT 1;

-- name: GetMany :many
SELECT * FROM users
ORDER BY username;

-- name: Delete :exec
DELETE FROM users
WHERE id = $1;

-- name: Insert :one
INSERT INTO users (name, username, email, password)
VALUES ($1, $2, $3, $4)
RETURNING id;

-- name: Update :exec
UPDATE users
SET name = $1, username = $2, email = $3, password = $4, updated_at = CURRENT_TIMESTAMP
WHERE id = $5;

-- name: Exists :one
SELECT EXISTS (
    SELECT 1 FROM users
    WHERE id = $1
) AS exists;

-- name: Count :one
SELECT COUNT(*) FROM users;