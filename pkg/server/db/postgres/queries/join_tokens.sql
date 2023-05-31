-- name: CreateJoinToken :one
INSERT INTO join_tokens(token, expires_at, trust_domain_id)
VALUES ($1, $2, $3)
RETURNING *;

-- name: UpdateJoinToken :one
UPDATE join_tokens
SET used       = $2,
    updated_at = now()
WHERE id = $1
RETURNING *;

-- name: DeleteJoinToken :exec
DELETE
FROM join_tokens
WHERE id = $1;

-- name: FindJoinTokenByID :one
SELECT *
FROM join_tokens
WHERE id = $1;

-- name: FindJoinToken :one
SELECT *
FROM join_tokens
WHERE token = $1;

-- name: FindJoinTokensByTrustDomainID :many
SELECT *
FROM join_tokens
WHERE trust_domain_id = $1;

-- name: ListJoinTokens :many
SELECT *
FROM join_tokens
ORDER BY created_at DESC;
