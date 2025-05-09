-- name: CreateChirp :one
INSERT INTO chirps (id, created_at, updated_at, body, user_id)
VALUES (
    gen_random_uuid(),
    NOW() AT TIME ZONE 'utc',
    NOW() AT TIME ZONE 'utc',
    $1,
    $2
)
RETURNING *;
--

-- name: GetAllChirps :many
SELECT * from chirps
ORDER BY created_at ASC;
--

-- name: GetChirpByID :one
SELECT * from chirps
WHERE id = $1;
--
