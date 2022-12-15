-- name: CreateFederationGroup :one
INSERT INTO federation_groups(name, description, status)
VALUES ($1, $2, $3)
RETURNING *;

-- name: DeleteFederationGroup :exec
DELETE
FROM federation_groups
WHERE id = $1;

-- name: UpdateFederationGroup :one
UPDATE federation_groups
SET name        = $2,
    description = $3,
    status       = $4
WHERE id = $1
RETURNING *;

-- name: FindFederationGroupByID :one
SELECT *
FROM federation_groups
WHERE id = $1;

-- name: ListFederationGroups :many
SELECT *
FROM federation_groups
ORDER BY created_at DESC;
