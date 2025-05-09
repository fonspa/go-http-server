-- name: CreateUser :one
INSERT INTO users (id, created_at, updated_at, email, hashed_password)
VALUES (
    gen_random_uuid(),
    NOW() AT TIME ZONE 'utc',
    NOW() AT TIME ZONE 'utc',
    $1,
    $2
)
RETURNING *;
--

-- name: DeleteUsers :exec
DELETE FROM users;
--

-- name: GetUserByEmail :one
SELECT * FROM users
WHERE email = $1;
--

-- name: UpdateUserCredentials :one
UPDATE users
SET email = $1, hashed_password = $2, updated_at = NOW() AT TIME ZONE 'utc'
WHERE id = $3
RETURNING *;
--
