// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.17.0
// source: relationships.sql

package sqlite

import (
	"context"
)

const createRelationship = `-- name: CreateRelationship :one
INSERT INTO relationships(id, trust_domain_a_id, trust_domain_b_id)
VALUES (?, ?, ?)
RETURNING id, trust_domain_a_id, trust_domain_b_id, trust_domain_a_consent, trust_domain_b_consent, created_at, updated_at
`

type CreateRelationshipParams struct {
	ID             string
	TrustDomainAID string
	TrustDomainBID string
}

func (q *Queries) CreateRelationship(ctx context.Context, arg CreateRelationshipParams) (Relationship, error) {
	row := q.queryRow(ctx, q.createRelationshipStmt, createRelationship, arg.ID, arg.TrustDomainAID, arg.TrustDomainBID)
	var i Relationship
	err := row.Scan(
		&i.ID,
		&i.TrustDomainAID,
		&i.TrustDomainBID,
		&i.TrustDomainAConsent,
		&i.TrustDomainBConsent,
		&i.CreatedAt,
		&i.UpdatedAt,
	)
	return i, err
}

const deleteRelationship = `-- name: DeleteRelationship :exec
DELETE
FROM relationships
WHERE id = ?
`

func (q *Queries) DeleteRelationship(ctx context.Context, id string) error {
	_, err := q.exec(ctx, q.deleteRelationshipStmt, deleteRelationship, id)
	return err
}

const findRelationshipByID = `-- name: FindRelationshipByID :one
SELECT id, trust_domain_a_id, trust_domain_b_id, trust_domain_a_consent, trust_domain_b_consent, created_at, updated_at
FROM relationships
WHERE id = ?
`

func (q *Queries) FindRelationshipByID(ctx context.Context, id string) (Relationship, error) {
	row := q.queryRow(ctx, q.findRelationshipByIDStmt, findRelationshipByID, id)
	var i Relationship
	err := row.Scan(
		&i.ID,
		&i.TrustDomainAID,
		&i.TrustDomainBID,
		&i.TrustDomainAConsent,
		&i.TrustDomainBConsent,
		&i.CreatedAt,
		&i.UpdatedAt,
	)
	return i, err
}

const findRelationshipsByTrustDomainID = `-- name: FindRelationshipsByTrustDomainID :many
SELECT id, trust_domain_a_id, trust_domain_b_id, trust_domain_a_consent, trust_domain_b_consent, created_at, updated_at
FROM relationships
WHERE trust_domain_a_id = ? OR trust_domain_b_id = ?
`

type FindRelationshipsByTrustDomainIDParams struct {
	TrustDomainAID string
	TrustDomainBID string
}

func (q *Queries) FindRelationshipsByTrustDomainID(ctx context.Context, arg FindRelationshipsByTrustDomainIDParams) ([]Relationship, error) {
	rows, err := q.query(ctx, q.findRelationshipsByTrustDomainIDStmt, findRelationshipsByTrustDomainID, arg.TrustDomainAID, arg.TrustDomainBID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []Relationship
	for rows.Next() {
		var i Relationship
		if err := rows.Scan(
			&i.ID,
			&i.TrustDomainAID,
			&i.TrustDomainBID,
			&i.TrustDomainAConsent,
			&i.TrustDomainBConsent,
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

const listRelationships = `-- name: ListRelationships :many
SELECT id, trust_domain_a_id, trust_domain_b_id, trust_domain_a_consent, trust_domain_b_consent, created_at, updated_at
FROM relationships
ORDER BY created_at DESC
`

func (q *Queries) ListRelationships(ctx context.Context) ([]Relationship, error) {
	rows, err := q.query(ctx, q.listRelationshipsStmt, listRelationships)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []Relationship
	for rows.Next() {
		var i Relationship
		if err := rows.Scan(
			&i.ID,
			&i.TrustDomainAID,
			&i.TrustDomainBID,
			&i.TrustDomainAConsent,
			&i.TrustDomainBConsent,
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

const updateRelationship = `-- name: UpdateRelationship :one
UPDATE relationships
SET trust_domain_a_consent = ?,
    trust_domain_b_consent = ?,
    updated_at = CURRENT_TIMESTAMP
WHERE id = ?
RETURNING id, trust_domain_a_id, trust_domain_b_id, trust_domain_a_consent, trust_domain_b_consent, created_at, updated_at
`

type UpdateRelationshipParams struct {
	TrustDomainAConsent string
	TrustDomainBConsent string
	ID                  string
}

func (q *Queries) UpdateRelationship(ctx context.Context, arg UpdateRelationshipParams) (Relationship, error) {
	row := q.queryRow(ctx, q.updateRelationshipStmt, updateRelationship, arg.TrustDomainAConsent, arg.TrustDomainBConsent, arg.ID)
	var i Relationship
	err := row.Scan(
		&i.ID,
		&i.TrustDomainAID,
		&i.TrustDomainBID,
		&i.TrustDomainAConsent,
		&i.TrustDomainBConsent,
		&i.CreatedAt,
		&i.UpdatedAt,
	)
	return i, err
}
