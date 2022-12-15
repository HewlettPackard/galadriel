-- name: CreateHarvester :one
INSERT INTO harvesters(member_id, is_leader, leader_until)
VALUES ($1, $2, $3)
RETURNING *;

-- name: UpdateHarvester :one
UPDATE harvesters
SET is_leader  = $2,
    leader_until = $3,
    updated_at = now()
WHERE id = $1
RETURNING *;

-- name: DeleteHarvester :exec
DELETE
FROM harvesters
WHERE id = $1;

-- name: FindHarvesterByID :one
SELECT *
FROM harvesters
WHERE id = $1;

-- name: FindHarvestersByMemberID :many
SELECT *
FROM harvesters
WHERE member_id = $1;

-- name: ListHarvesters :many
SELECT *
FROM harvesters
ORDER BY member_id;