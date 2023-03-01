-- name: CreateBundle :one
INSERT INTO bundles(data, digest, signature, digest_algorithm, signature_algorithm, signing_cert, trust_domain_id)
VALUES ($1, $2, $3, $4, $5, $6, $7)
RETURNING *;

-- name: UpdateBundle :one
UPDATE bundles
SET data                = $2,
    digest              = $3,
    signature           = $4,
    digest_algorithm    = $5,
    signature_algorithm = $6,
    signing_cert        = $7,
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
