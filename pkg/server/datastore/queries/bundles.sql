-- name: CreateBundle :one
INSERT INTO bundles(data, signature, signature_algorithm, signing_certificate, trust_domain_id)
VALUES ($1, $2, $3, $4, $5)
RETURNING *;

-- name: UpdateBundle :one
UPDATE bundles
SET data                = $2,
    signature           = $3,
    signature_algorithm = $4,
    signing_certificate = $5,
    updated_at          = now()
WHERE id = $1
RETURNING *;

-- name: DeleteBundle :exec
DELETE
FROM bundles
WHERE id = $1;

-- name: FindBundleByID :one
SELECT *
FROM bundles
WHERE id = $1;

-- name: FindBundleByTrustDomainID :one
SELECT *
FROM bundles
WHERE trust_domain_id = $1;

-- name: ListBundles :many
SELECT *
FROM bundles
ORDER BY created_at DESC;
