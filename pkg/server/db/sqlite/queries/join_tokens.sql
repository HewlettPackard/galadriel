-- name: CreateJoinToken :one
INSERT INTO join_tokens(id, token, expires_at, trust_domain_id)
VALUES (?, ?, ?, ?)
RETURNING *;

-- name: UpdateJoinToken :one
UPDATE join_tokens
SET used       = ?,
    updated_at = datetime('now')
WHERE id = ?
RETURNING *;

-- name: DeleteJoinToken :exec
DELETE
FROM join_tokens
WHERE id = ?;

-- name: FindJoinTokenByID :one
SELECT *
FROM join_tokens
WHERE id = ?;

-- name: FindJoinToken :one
SELECT *
FROM join_tokens
WHERE token = ?;

-- name: FindJoinTokensByTrustDomainID :many
SELECT *
FROM join_tokens
WHERE trust_domain_id = ?
ORDER BY created_at DESC;

-- name: ListJoinTokens :many
SELECT *
FROM join_tokens
ORDER BY created_at DESC;
