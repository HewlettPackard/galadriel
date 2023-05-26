-- name: CreateTrustDomain :one
INSERT INTO trust_domains(name, description)
VALUES ($1, $2)
RETURNING *;

-- name: UpdateTrustDomain :one
UPDATE trust_domains
SET description = $2,
    updated_at  = now()
WHERE id = $1
RETURNING *;

-- name: DeleteTrustDomain :exec
DELETE
FROM trust_domains
WHERE id = $1;

-- name: FindTrustDomainByID :one
SELECT *
FROM trust_domains
WHERE id = $1;

-- name: FindTrustDomainByName :one
SELECT *
FROM trust_domains
WHERE name = $1;

-- name: ListTrustDomains :many
SELECT *
FROM trust_domains
ORDER BY name;
