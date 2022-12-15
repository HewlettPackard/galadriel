-- name: CreateMembership :one
INSERT INTO memberships(member_id, federation_group_id, status)
VALUES ($1, $2, $3)
RETURNING *;

-- name: UpdateMembership :one
UPDATE memberships
SET status     = $2,
    updated_at = now()
WHERE id = $1
RETURNING *;

-- name: DeleteMembership :exec
DELETE
FROM memberships
WHERE id = $1;

-- name: FindMembershipByID :one
SELECT *
FROM memberships
WHERE id = $1;

-- name: FindMembershipsByMemberID :many
SELECT *
FROM memberships
WHERE member_id = $1;

-- name: ListMemberships :many
SELECT *
FROM memberships
ORDER BY created_at DESC;
