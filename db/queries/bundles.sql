-- name: CreateBundle :one
INSERT INTO bundles(raw_bundle, digest, signed_bundle, tlog_id, svid_pem, member_id)
VALUES ($1, $2, $3, $4, $5, $6)
RETURNING *;

-- name: UpdateBundle :one
UPDATE bundles
SET raw_bundle    = $2,
    digest        = $3,
    signed_bundle = $4,
    tlog_id       = $5,
    svid_pem      = $6,
    updated_at    = now()
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

-- name: FindBundleByMemberID :one
SELECT *
FROM bundles
WHERE member_id = $1;

-- name: ListBundles :many
SELECT *
FROM bundles
ORDER BY created_at DESC;
