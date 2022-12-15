-- name: CreateMember :one
INSERT INTO members(trust_domain, status)
VALUES ($1, $2)
RETURNING *;

-- name: UpdateMember :one
UPDATE members
SET trust_domain = $2,
    status     = $3,
    updated_at = now()
WHERE id = $1
RETURNING *;

-- name: DeleteMember :exec
DELETE
FROM members
WHERE id = $1;

-- name: FindMemberByID :one
SELECT *
FROM members
WHERE id = $1;

-- name: FindMemberByTrustDomain :one
SELECT *
FROM members
WHERE trust_domain = $1;

-- name: ListMembers :many
SELECT *
FROM members
ORDER BY trust_domain;
