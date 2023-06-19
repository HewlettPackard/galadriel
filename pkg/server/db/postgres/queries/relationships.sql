-- name: CreateRelationship :one
INSERT INTO relationships(trust_domain_a_id, trust_domain_b_id, trust_domain_a_consent, trust_domain_b_consent, created_at)
VALUES ($1, $2, $3, $4, $5)
RETURNING *;

-- name: UpdateRelationship :one
UPDATE relationships
SET trust_domain_a_consent = $2,
    trust_domain_b_consent = $3,
    updated_at             = now()
WHERE id = $1
RETURNING *;

-- name: DeleteRelationship :exec
DELETE
FROM relationships
WHERE id = $1;

-- name: FindRelationshipByID :one
SELECT *
FROM relationships
WHERE id = $1;

-- name: FindRelationshipsByTrustDomainID :many
SELECT *
FROM relationships
WHERE trust_domain_a_id = $1
   OR trust_domain_b_id = $1;

