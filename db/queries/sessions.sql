-- name: CreateSession :exec
INSERT INTO sessions (
    token_hash, user_id, ip_address, user_agent, expires_at
) VALUES (
    ?, ?, ?, ?, ?
);

-- name: GetSession :one
SELECT 
    *
FROM sessions s
JOIN users u ON s.user_id = u.id
WHERE s.token_hash = ? AND s.expires_at > datetime('now')
LIMIT 1;

-- name: DeleteSession :exec
DELETE FROM sessions
WHERE token_hash = ?;

-- name: DeleteAllUserSessions :exec
-- Útil cuando un usuario cambia su password o es desactivado
DELETE FROM sessions
WHERE user_id = ?;

-- name: CleanExpiredSessions :exec
-- Esto lo puedes correr en un ticker de Go cada 1 hora
DELETE FROM sessions
WHERE expires_at < datetime('now');