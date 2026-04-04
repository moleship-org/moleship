-- +goose Up

-- +goose StatementBegin
CREATE TABLE users (
    id BLOB PRIMARY KEY NOT NULL,
    username TEXT NOT NULL UNIQUE,
    first_name TEXT,
    last_name TEXT,
    password_hash TEXT NOT NULL,
    email TEXT NOT NULL UNIQUE,
    is_admin BOOLEAN NOT NULL DEFAULT FALSE,
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    last_login TEXT,
    created_at TEXT NOT NULL DEFAULT (datetime('now')),
    updated_at TEXT NOT NULL DEFAULT (datetime('now')),
    deleted_at TEXT
);
-- +goose StatementEnd

-- +goose StatementBegin
CREATE TRIGGER IF NOT EXISTS tr_users_updated_at 
AFTER UPDATE ON users
BEGIN
    UPDATE users SET updated_at = datetime('now') WHERE id = OLD.id;
END;
-- +goose StatementEnd

-- +goose Down
DROP TRIGGER IF EXISTS tr_users_updated_at;
DROP TABLE users;