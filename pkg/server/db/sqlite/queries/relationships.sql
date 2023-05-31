-- name: CreateRelationship :one
INSERT INTO relationships(id, trust_domain_a_id, trust_domain_b_id)
VALUES (?, ?, ?)
RETURNING *;

-- name: UpdateRelationship :one
UPDATE relationships
SET trust_domain_a_consent = ?,
    trust_domain_b_consent = ?,
    updated_at             = CURRENT_TIMESTAMP
WHERE id = ?
RETURNING *;

-- name: DeleteRelationship :exec
DELETE
FROM relationships
WHERE id = ?;

-- name: FindRelationshipByID :one
SELECT *
FROM relationships
WHERE id = ?;

-- name: FindRelationshipsByTrustDomainID :many
SELECT *
FROM relationships
WHERE trust_domain_a_id = ?
   OR trust_domain_b_id = ?;

-- name: ListRelationships :many
SELECT *
FROM relationships
ORDER BY created_at DESC;
