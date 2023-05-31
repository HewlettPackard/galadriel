-- name: CreateTrustDomain :one
INSERT INTO trust_domains(id, name, description)
VALUES (?, ?, ?)
RETURNING *;

-- name: UpdateTrustDomain :one
UPDATE trust_domains
SET description = ?,
    updated_at  = datetime('now')
WHERE id = ?
RETURNING *;

-- name: DeleteTrustDomain :exec
DELETE
FROM trust_domains
WHERE id = ?;

-- name: FindTrustDomainByID :one
SELECT *
FROM trust_domains
WHERE id = ?;

-- name: FindTrustDomainByName :one
SELECT *
FROM trust_domains
WHERE name = ?;

-- name: ListTrustDomains :many
SELECT *
FROM trust_domains
ORDER BY name;
