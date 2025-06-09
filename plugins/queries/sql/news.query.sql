-- name: Get :one
SELECT * FROM news
WHERE id = $1 LIMIT 1;

-- name: GetMany :many
SELECT * FROM news
ORDER BY category;

-- name: Create :one
INSERT INTO news (
  content, category
) VALUES (
  $1, $2
)
RETURNING *;

-- name: Update :one
UPDATE news
  set content = $2,
  category = $3
WHERE id = $1
RETURNING *;

-- name: Delete :exec
DELETE FROM news
WHERE id = $1;