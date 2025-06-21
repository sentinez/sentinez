-- name: Get :one
SELECT * FROM authors
WHERE id = $1 LIMIT 1;

-- name: GetMany :many
SELECT * FROM authors
ORDER BY name;

-- name: Create :one
INSERT INTO authors (
  name, bio
) VALUES (
  $1, $2
)
RETURNING *;

-- name: Update :one
UPDATE authors
  set name = $2,
  bio = $3
WHERE id = $1
RETURNING *;

-- name: Delete :exec
DELETE FROM authors
WHERE id = $1;