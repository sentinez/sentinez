-- name: Insert :exec
INSERT INTO dev_sentinez_accounts (id, data)
VALUES ($1, $2::jsonb);

-- name: GetByID :one
SELECT * FROM dev_sentinez_accounts
WHERE id = $1
LIMIT 1;

-- name: GetByUsernameOrEmail :one
SELECT * FROM dev_sentinez_accounts
WHERE data->>'username' = $1 OR data->>'email' = $1
LIMIT 1;

-- name: GetMany :many
SELECT * FROM dev_sentinez_accounts
ORDER BY data->>'username'
LIMIT $1 OFFSET $2;

-- name: Update :exec
UPDATE dev_sentinez_accounts
SET data = $2::jsonb,
    updated_at = CURRENT_TIMESTAMP
WHERE id = $1;

-- name: Delete :exec
DELETE FROM dev_sentinez_accounts
WHERE id = $1;
