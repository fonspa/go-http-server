-- +goose Up
ALTER TABLE users
ADD COLUMN IF NOT EXISTS is_chirpy_red BOOLEAN DEFAULT FALSE;

-- +goose Down
ALTER TABLE users
DROP COLUMN IF EXISTS is_chirpy_red;
