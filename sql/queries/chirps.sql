-- name: CreateChirp :one
INSERT INTO chirps (id, created_at, updated_at, body, user_id)
VALUES (
    $1,
    $2,
    $3,
    $4,
    $5
)
RETURNING *;

-- name: DeleteChirp :exec
DELETE FROM chirps WHERE id = $1;

-- name: DeleteAllChirps :exec
DELETE FROM chirps WHERE 1=1;

-- name: GetAllChirps :many
SELECT * FROM chirps ORDER BY created_at ASC;

-- name: GetChirpByID :one
SELECT * FROM chirps WHERE id = $1 LIMIT 1;

-- name: GetChirpsByUserID :many
SELECT * FROM chirps WHERE user_id = $1;

-- name: UpdateChirp :one
UPDATE chirps SET body = $2, updated_at = $3 WHERE id = $1 RETURNING *;