-- name: CreateRefreshToken :one
INSERT INTO refresh_tokens (token, created_at, updated_at, user_id, expires_at)
VALUES (
    $1,
    $2,
    $3,
    $4,
    $5
)
RETURNING *;

-- name: DeleteExpiredRefreshTokens :exec
DELETE FROM refresh_tokens WHERE expires_at < $1 AND revoked_at IS NULL;

-- name: DeleteRefreshToken :exec
DELETE FROM refresh_tokens WHERE token = $1;

-- name: DeleteAllRefreshTokens :exec
DELETE FROM refresh_tokens WHERE 1=1;

-- name: GetRefreshTokenByToken :one
SELECT * FROM refresh_tokens WHERE token = $1 LIMIT 1;

-- name: GetRefreshTokenByUserID :many
SELECT * FROM refresh_tokens WHERE user_id = $1;

-- name: RevokeRefreshToken :one
UPDATE refresh_tokens SET revoked_at = $2, updated_at = $3 WHERE token = $1 RETURNING *;