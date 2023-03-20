// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.17.0
// source: trust_domains.sql

package datastore

import (
	"context"
	"database/sql"

	"github.com/jackc/pgtype"
)

const createTrustDomain = `-- name: CreateTrustDomain :one
INSERT INTO trust_domains(name, description)
VALUES ($1, $2)
RETURNING id, name, description, harvester_spiffe_id, onboarding_bundle, created_at, updated_at
`

type CreateTrustDomainParams struct {
	Name        string
	Description sql.NullString
}

func (q *Queries) CreateTrustDomain(ctx context.Context, arg CreateTrustDomainParams) (TrustDomain, error) {
	row := q.queryRow(ctx, q.createTrustDomainStmt, createTrustDomain, arg.Name, arg.Description)
	var i TrustDomain
	err := row.Scan(
		&i.ID,
		&i.Name,
		&i.Description,
		&i.HarvesterSpiffeID,
		&i.OnboardingBundle,
		&i.CreatedAt,
		&i.UpdatedAt,
	)
	return i, err
}

const deleteTrustDomain = `-- name: DeleteTrustDomain :exec
DELETE
FROM trust_domains
WHERE id = $1
`

func (q *Queries) DeleteTrustDomain(ctx context.Context, id pgtype.UUID) error {
	_, err := q.exec(ctx, q.deleteTrustDomainStmt, deleteTrustDomain, id)
	return err
}

const findTrustDomainByID = `-- name: FindTrustDomainByID :one
SELECT id, name, description, harvester_spiffe_id, onboarding_bundle, created_at, updated_at
FROM trust_domains
WHERE id = $1
`

func (q *Queries) FindTrustDomainByID(ctx context.Context, id pgtype.UUID) (TrustDomain, error) {
	row := q.queryRow(ctx, q.findTrustDomainByIDStmt, findTrustDomainByID, id)
	var i TrustDomain
	err := row.Scan(
		&i.ID,
		&i.Name,
		&i.Description,
		&i.HarvesterSpiffeID,
		&i.OnboardingBundle,
		&i.CreatedAt,
		&i.UpdatedAt,
	)
	return i, err
}

const findTrustDomainByName = `-- name: FindTrustDomainByName :one
SELECT id, name, description, harvester_spiffe_id, onboarding_bundle, created_at, updated_at
FROM trust_domains
WHERE name = $1
`

func (q *Queries) FindTrustDomainByName(ctx context.Context, name string) (TrustDomain, error) {
	row := q.queryRow(ctx, q.findTrustDomainByNameStmt, findTrustDomainByName, name)
	var i TrustDomain
	err := row.Scan(
		&i.ID,
		&i.Name,
		&i.Description,
		&i.HarvesterSpiffeID,
		&i.OnboardingBundle,
		&i.CreatedAt,
		&i.UpdatedAt,
	)
	return i, err
}

const listTrustDomains = `-- name: ListTrustDomains :many
SELECT id, name, description, harvester_spiffe_id, onboarding_bundle, created_at, updated_at
FROM trust_domains
ORDER BY name
`

func (q *Queries) ListTrustDomains(ctx context.Context) ([]TrustDomain, error) {
	rows, err := q.query(ctx, q.listTrustDomainsStmt, listTrustDomains)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []TrustDomain
	for rows.Next() {
		var i TrustDomain
		if err := rows.Scan(
			&i.ID,
			&i.Name,
			&i.Description,
			&i.HarvesterSpiffeID,
			&i.OnboardingBundle,
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

const updateTrustDomain = `-- name: UpdateTrustDomain :one
UPDATE trust_domains
SET description         = $2,
    harvester_spiffe_id = $3,
    onboarding_bundle   = $4,
    updated_at          = now()
WHERE id = $1
RETURNING id, name, description, harvester_spiffe_id, onboarding_bundle, created_at, updated_at
`

type UpdateTrustDomainParams struct {
	ID                pgtype.UUID
	Description       sql.NullString
	HarvesterSpiffeID sql.NullString
	OnboardingBundle  []byte
}

func (q *Queries) UpdateTrustDomain(ctx context.Context, arg UpdateTrustDomainParams) (TrustDomain, error) {
	row := q.queryRow(ctx, q.updateTrustDomainStmt, updateTrustDomain,
		arg.ID,
		arg.Description,
		arg.HarvesterSpiffeID,
		arg.OnboardingBundle,
	)
	var i TrustDomain
	err := row.Scan(
		&i.ID,
		&i.Name,
		&i.Description,
		&i.HarvesterSpiffeID,
		&i.OnboardingBundle,
		&i.CreatedAt,
		&i.UpdatedAt,
	)
	return i, err
}
