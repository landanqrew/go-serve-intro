-- name: CreateUser :one
INSERT INTO users (id, created_at, updated_at, email, hashed_password)
VALUES (
    $1,
    $2,
    $3,
    $4,
    $5
)
RETURNING *;

-- name: DeleteUser :exec
DELETE FROM users WHERE id = $1;

-- name: DeleteAllUsers :exec
DELETE FROM users WHERE 1=1;

-- name: GetUsersByEmail :many
SELECT * FROM users WHERE email = $1;

-- name: GetUserByID :one
SELECT * FROM users WHERE id = $1 LIMIT 1;

-- name: UpdateUserPasswordByID :one
UPDATE users SET hashed_password = $2, updated_at = $3 WHERE id = $1 RETURNING *;

-- name: UpdateUser :one
UPDATE users SET email = $2, hashed_password = $3, updated_at = $4 WHERE id = $1 RETURNING *;

-- name: UpdateUserEmailByID :one
UPDATE users SET email = $2, updated_at = $3 WHERE id = $1 RETURNING *;

-- name: UpdateUserPasswordByEmail :one
UPDATE users SET hashed_password = $2 WHERE email = $1 RETURNING *;

-- name: UpdateUserSetChirpyRed :one
UPDATE users SET is_chirpy_red = TRUE, updated_at = $2 WHERE id = $1 RETURNING *;

-- name: UpdateUserUnsetChirpyRed :one
UPDATE users SET is_chirpy_red = FALSE, updated_at = $2 WHERE id = $1 RETURNING *;