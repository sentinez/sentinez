-- name: GetByID :one
SELECT *
FROM users
WHERE id = $1
LIMIT 1;

-- name: GetByUsernameOrEmail :one
SELECT *
FROM users
WHERE data->>'username' = $1
   OR data->>'email' = $1
LIMIT 1;

-- name: GetPage :many
SELECT *
FROM users
ORDER BY data->>'username'
OFFSET $1
LIMIT $2;

-- name: Delete :exec
DELETE FROM users
WHERE id = $1;

-- name: Insert :one
INSERT INTO users (id, data)
VALUES ($1, $2::jsonb)
RETURNING id;

-- name: Update :exec
UPDATE users
SET
    data = $1::jsonb,
    updated_at = CURRENT_TIMESTAMP
WHERE id = $2;

-- name: ExistsByUsername :one
SELECT EXISTS (
    SELECT 1
    FROM users
    WHERE data->>'username' = $1
) AS exists;

-- name: ExistsByEmail :one
SELECT EXISTS (
    SELECT 1
    FROM users
    WHERE data->>'email' = $1
) AS exists;

-- name: Count :one
SELECT COUNT(*) AS count
FROM users;
