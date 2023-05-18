-- name: CreateBundle :one
INSERT INTO bundles(id, data, digest, signature, signing_certificate, trust_domain_id)
VALUES (?, ?, ?, ?, ?, ?)
RETURNING *;

-- name: UpdateBundle :one
UPDATE bundles
SET data                = ?,
    digest              = ?,
    signature           = ?,
    signing_certificate = ?,
    updated_at          = datetime('now')
WHERE id = ?
RETURNING *;

-- name: DeleteBundle :exec
DELETE
FROM bundles
WHERE id = ?;

-- name: FindBundleByID :one
SELECT *
FROM bundles
WHERE id = ?;

-- name: FindBundleByTrustDomainID :one
SELECT *
FROM bundles
WHERE trust_domain_id = ?
LIMIT 1;

-- name: ListBundles :many
SELECT *
FROM bundles
ORDER BY created_at DESC;
