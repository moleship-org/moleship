-- name: GetUser :one
SELECT * FROM users
WHERE id = ? LIMIT 1;

-- name: GetUserByUsername :one
SELECT * FROM users
WHERE username = ? LIMIT 1;

-- name: GetUserByEmail :one
SELECT * FROM users
WHERE email = ? LIMIT 1;

-- name: ListUsers :many
SELECT * FROM users
WHERE deleted_at IS NULL
ORDER BY created_at DESC
LIMIT ? OFFSET ?;

-- name: CountUsers :one
SELECT count(*) FROM users;

-- name: CreateUser :exec
INSERT INTO users (
    id, username, first_name, last_name, password_hash, email, is_admin
) VALUES (
    ?, ?, ?, ?, ?, ?, ?
);

-- name: UpdateUser :exec
UPDATE users
SET 
    username = ?,
    first_name = ?,
    last_name = ?,
    password_hash = ?,
    email = ?,
    is_admin = ?,
    is_active = ?
WHERE id = ?;

-- name: UpdateUserLastLogin :exec
UPDATE users
SET last_login = datetime('now')
WHERE id = ?;

-- name: ActivateUser :exec
UPDATE users
SET is_active = 1
WHERE id = ?;

-- name: DeactivateUser :exec
UPDATE users
SET is_active = 0
WHERE id = ?;

-- name: SoftDeleteUser :exec
UPDATE users
SET deleted_at = datetime('now'),
    is_active = 0
WHERE id = ?;

-- name: HardDeleteUser :exec
DELETE FROM users
WHERE id = ?;