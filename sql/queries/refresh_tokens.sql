-- name: CreateRefreshToken :one
INSERT INTO refresh_tokens(token, created_at, updated_at, user_id, expires_at, revoked_at)
VALUES (
    $1,
    NOW() AT TIME ZONE 'utc',
    NOW() AT TIME ZONE 'utc',
    $2,
    $3,
    NULL
)
RETURNING *;
--

-- name: GetUserFromRefreshToken :one
SELECT * FROM refresh_tokens
WHERE token = $1;
--

-- name: RevokeRefreshToken :exec
UPDATE refresh_tokens
SET
    updated_at = NOW() AT TIME ZONE 'utc',
    revoked_at = NOW() AT TIME ZONE 'utc'
WHERE token = $1;
--
