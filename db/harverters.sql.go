// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.16.0
// source: harverters.sql

package db

import (
	"context"
	"database/sql"
	"time"

	"github.com/jackc/pgtype"
)

const createHarvester = `-- name: CreateHarvester :one
INSERT INTO harvesters(member_id, is_leader, leader_until)
VALUES ($1, $2, $3)
RETURNING id, member_id, is_leader, leader_until, created_at, updated_at
`

type CreateHarvesterParams struct {
	MemberID    pgtype.UUID
	IsLeader    sql.NullBool
	LeaderUntil time.Time
}

func (q *Queries) CreateHarvester(ctx context.Context, arg CreateHarvesterParams) (Harvester, error) {
	row := q.queryRow(ctx, q.createHarvesterStmt, createHarvester, arg.MemberID, arg.IsLeader, arg.LeaderUntil)
	var i Harvester
	err := row.Scan(
		&i.ID,
		&i.MemberID,
		&i.IsLeader,
		&i.LeaderUntil,
		&i.CreatedAt,
		&i.UpdatedAt,
	)
	return i, err
}

const deleteHarvester = `-- name: DeleteHarvester :exec
DELETE
FROM harvesters
WHERE id = $1
`

func (q *Queries) DeleteHarvester(ctx context.Context, id pgtype.UUID) error {
	_, err := q.exec(ctx, q.deleteHarvesterStmt, deleteHarvester, id)
	return err
}

const findHarvesterByID = `-- name: FindHarvesterByID :one
SELECT id, member_id, is_leader, leader_until, created_at, updated_at
FROM harvesters
WHERE id = $1
`

func (q *Queries) FindHarvesterByID(ctx context.Context, id pgtype.UUID) (Harvester, error) {
	row := q.queryRow(ctx, q.findHarvesterByIDStmt, findHarvesterByID, id)
	var i Harvester
	err := row.Scan(
		&i.ID,
		&i.MemberID,
		&i.IsLeader,
		&i.LeaderUntil,
		&i.CreatedAt,
		&i.UpdatedAt,
	)
	return i, err
}

const findHarvestersByMemberID = `-- name: FindHarvestersByMemberID :many
SELECT id, member_id, is_leader, leader_until, created_at, updated_at
FROM harvesters
WHERE member_id = $1
`

func (q *Queries) FindHarvestersByMemberID(ctx context.Context, memberID pgtype.UUID) ([]Harvester, error) {
	rows, err := q.query(ctx, q.findHarvestersByMemberIDStmt, findHarvestersByMemberID, memberID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []Harvester
	for rows.Next() {
		var i Harvester
		if err := rows.Scan(
			&i.ID,
			&i.MemberID,
			&i.IsLeader,
			&i.LeaderUntil,
			&i.CreatedAt,
			&i.UpdatedAt,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Close(); err != nil {
		return nil, err
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const listHarvesters = `-- name: ListHarvesters :many
SELECT id, member_id, is_leader, leader_until, created_at, updated_at
FROM harvesters
ORDER BY member_id
`

func (q *Queries) ListHarvesters(ctx context.Context) ([]Harvester, error) {
	rows, err := q.query(ctx, q.listHarvestersStmt, listHarvesters)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []Harvester
	for rows.Next() {
		var i Harvester
		if err := rows.Scan(
			&i.ID,
			&i.MemberID,
			&i.IsLeader,
			&i.LeaderUntil,
			&i.CreatedAt,
			&i.UpdatedAt,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Close(); err != nil {
		return nil, err
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const updateHarvester = `-- name: UpdateHarvester :one
UPDATE harvesters
SET is_leader  = $2,
    leader_until = $3,
    updated_at = now()
WHERE id = $1
RETURNING id, member_id, is_leader, leader_until, created_at, updated_at
`

type UpdateHarvesterParams struct {
	ID          pgtype.UUID
	IsLeader    sql.NullBool
	LeaderUntil time.Time
}

func (q *Queries) UpdateHarvester(ctx context.Context, arg UpdateHarvesterParams) (Harvester, error) {
	row := q.queryRow(ctx, q.updateHarvesterStmt, updateHarvester, arg.ID, arg.IsLeader, arg.LeaderUntil)
	var i Harvester
	err := row.Scan(
		&i.ID,
		&i.MemberID,
		&i.IsLeader,
		&i.LeaderUntil,
		&i.CreatedAt,
		&i.UpdatedAt,
	)
	return i, err
}
