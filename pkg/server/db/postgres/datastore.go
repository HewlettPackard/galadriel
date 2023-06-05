package postgres

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/HewlettPackard/galadriel/pkg/common/entity"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/stdlib"
	"github.com/pkg/errors"
	"github.com/spiffe/go-spiffe/v2/spiffeid"
)

// Datastore is a SQL database accessor that provides convenient methods
// to perform CRUD operations for Galadriel entities.
// It implements the Datastore interface.
type Datastore struct {
	db      *sql.DB
	querier Querier
}

// NewDatastore creates a new instance of a Datastore object that connects to a Postgres database
// parsing the connString.
// The connString can be a URL, e.g, "postgresql://host...", or a DSN, e.g., "host= user= password= dbname= port=".
func NewDatastore(connString string) (*Datastore, error) {
	c, err := pgx.ParseConfig(connString)
	if err != nil {
		return nil, fmt.Errorf("failed to parse Postgres Connection URL: %w", err)
	}

	db := stdlib.OpenDB(*c)

	// validates if the schema in the DB matches the schema supported by the app, and runs the migrations if needed
	if err = validateAndMigrateSchema(db); err != nil {
		return nil, fmt.Errorf("failed to validate or migrate schema: %w", err)
	}

	return &Datastore{
		db:      db,
		querier: New(db),
	}, nil
}

func (d *Datastore) Close() error {
	return d.db.Close()
}

// CreateOrUpdateTrustDomain creates or updates the given TrustDomain in the underlying db, based on
// whether the given entity has an ID, in which case, it is updated.
func (d *Datastore) CreateOrUpdateTrustDomain(ctx context.Context, req *entity.TrustDomain) (*entity.TrustDomain, error) {
	if req.Name.String() == "" {
		return nil, errors.New("trustDomain trust domain is missing")
	}

	var trustDomain *TrustDomain
	var err error
	if req.ID.Valid {
		trustDomain, err = d.updateTrustDomain(ctx, req)
	} else {
		trustDomain, err = d.createTrustDomain(ctx, req)
	}
	if err != nil {
		return nil, err
	}

	response, err := trustDomain.ToEntity()
	if err != nil {
		return nil, fmt.Errorf("failed converting trustDomain model to entity: %w", err)
	}

	return response, nil
}

func (d *Datastore) DeleteTrustDomain(ctx context.Context, trustDomainID uuid.UUID) error {
	pgID, err := uuidToPgType(trustDomainID)
	if err != nil {
		return err
	}

	if err = d.querier.DeleteTrustDomain(ctx, pgID); err != nil {
		return fmt.Errorf("failed deleting trust domain with ID=%q: %w", trustDomainID, err)
	}

	return nil
}

func (d *Datastore) ListTrustDomains(ctx context.Context) ([]*entity.TrustDomain, error) {
	trustDomains, err := d.querier.ListTrustDomains(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed getting trust domain list: %w", err)
	}

	result := make([]*entity.TrustDomain, len(trustDomains))
	for i, m := range trustDomains {
		r, err := m.ToEntity()
		if err != nil {
			return nil, fmt.Errorf("failed converting model trust domain to entity: %v", err)
		}
		result[i] = r
	}

	return result, nil
}

func (d *Datastore) FindTrustDomainByID(ctx context.Context, trustDomainID uuid.UUID) (*entity.TrustDomain, error) {
	pgID, err := uuidToPgType(trustDomainID)
	if err != nil {
		return nil, err
	}

	m, err := d.querier.FindTrustDomainByID(ctx, pgID)
	switch {
	case errors.Is(err, sql.ErrNoRows):
		return nil, nil
	case err != nil:
		return nil, fmt.Errorf("failed looking up trust domain for ID=%q: %w", trustDomainID, err)
	}

	r, err := m.ToEntity()
	if err != nil {
		return nil, fmt.Errorf("failed converting model trust domain to entity: %w", err)
	}

	return r, nil
}

func (d *Datastore) FindTrustDomainByName(ctx context.Context, name spiffeid.TrustDomain) (*entity.TrustDomain, error) {
	trustDomain, err := d.querier.FindTrustDomainByName(ctx, name.String())
	switch {
	case errors.Is(err, sql.ErrNoRows):
		return nil, nil
	case err != nil:
		return nil, fmt.Errorf("failed looking up trust domain for Trust Domain=%q: %w", name, err)
	}

	r, err := trustDomain.ToEntity()
	if err != nil {
		return nil, fmt.Errorf("failed converting model trust domain to entity: %w", err)
	}

	return r, nil
}

func (d *Datastore) CreateOrUpdateBundle(ctx context.Context, req *entity.Bundle) (*entity.Bundle, error) {
	var bundle *Bundle
	var err error
	if req.ID.Valid {
		bundle, err = d.updateBundle(ctx, req)
	} else {
		bundle, err = d.createBundle(ctx, req)
	}
	if err != nil {
		return nil, err
	}

	response, err := bundle.ToEntity()
	if err != nil {
		return nil, fmt.Errorf("failed converting trust domain model to entity: %w", err)
	}

	return response, nil
}

func (d *Datastore) FindBundleByID(ctx context.Context, bundleID uuid.UUID) (*entity.Bundle, error) {
	pgID, err := uuidToPgType(bundleID)
	if err != nil {
		return nil, err
	}

	bundle, err := d.querier.FindBundleByID(ctx, pgID)
	switch {
	case errors.Is(err, sql.ErrNoRows):
		return nil, nil
	case err != nil:
		return nil, fmt.Errorf("failed looking up bundle with ID=%q: %w", bundleID, err)
	}

	b, err := bundle.ToEntity()
	if err != nil {
		return nil, fmt.Errorf("failed converting model bundle to entity: %w", err)
	}

	return b, nil
}

func (d *Datastore) FindBundleByTrustDomainID(ctx context.Context, trustDomainID uuid.UUID) (*entity.Bundle, error) {
	pgID, err := uuidToPgType(trustDomainID)
	if err != nil {
		return nil, err
	}

	trustDomain, err := d.querier.FindBundleByTrustDomainID(ctx, pgID)
	switch {
	case errors.Is(err, sql.ErrNoRows):
		return nil, nil
	case err != nil:
		return nil, fmt.Errorf("failed looking up bundle for ID=%q: %w", trustDomainID, err)
	}

	td, err := trustDomain.ToEntity()
	if err != nil {
		return nil, fmt.Errorf("failed converting model bundle to entity: %w", err)
	}

	return td, nil
}

func (d *Datastore) ListBundles(ctx context.Context) ([]*entity.Bundle, error) {
	bundles, err := d.querier.ListBundles(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed getting bundle list: %w", err)
	}

	result := make([]*entity.Bundle, len(bundles))
	for i, m := range bundles {
		r, err := m.ToEntity()
		if err != nil {
			return nil, fmt.Errorf("failed converting model bundle to entity: %w", err)
		}
		result[i] = r
	}

	return result, nil
}

func (d *Datastore) DeleteBundle(ctx context.Context, bundleID uuid.UUID) error {
	pgID, err := uuidToPgType(bundleID)
	if err != nil {
		return err
	}

	if err = d.querier.DeleteBundle(ctx, pgID); err != nil {
		return fmt.Errorf("failed deleting bundle with ID=%q: %w", bundleID, err)
	}

	return nil
}

func (d *Datastore) CreateJoinToken(ctx context.Context, req *entity.JoinToken) (*entity.JoinToken, error) {
	pgID, err := uuidToPgType(req.TrustDomainID)
	if err != nil {
		return nil, err
	}

	params := CreateJoinTokenParams{
		Token:         req.Token,
		ExpiresAt:     req.ExpiresAt,
		TrustDomainID: pgID,
	}
	joinToken, err := d.querier.CreateJoinToken(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("failed creating join token: %w", err)
	}

	return joinToken.ToEntity(), nil
}

func (d *Datastore) FindJoinTokensByID(ctx context.Context, joinTokenID uuid.UUID) (*entity.JoinToken, error) {
	pgID, err := uuidToPgType(joinTokenID)
	if err != nil {
		return nil, err
	}

	joinToken, err := d.querier.FindJoinTokenByID(ctx, pgID)
	switch {
	case errors.Is(err, sql.ErrNoRows):
		return nil, nil
	case err != nil:
		return nil, fmt.Errorf("failed looking up join token with ID=%q: %w", joinTokenID, err)
	}

	return joinToken.ToEntity(), nil
}

func (d *Datastore) FindJoinTokensByTrustDomainID(ctx context.Context, trustDomainID uuid.UUID) ([]*entity.JoinToken, error) {
	pgID, err := uuidToPgType(trustDomainID)
	if err != nil {
		return nil, err
	}

	tokens, err := d.querier.FindJoinTokensByTrustDomainID(ctx, pgID)
	switch {
	case errors.Is(err, sql.ErrNoRows):
		return nil, nil
	case err != nil:
		return nil, fmt.Errorf("failed looking up join token for Name ID=%q: %w", trustDomainID, err)
	}

	result := make([]*entity.JoinToken, len(tokens))
	for i, t := range tokens {
		result[i] = t.ToEntity()
	}

	return result, nil
}

func (d *Datastore) ListJoinTokens(ctx context.Context) ([]*entity.JoinToken, error) {
	tokens, err := d.querier.ListJoinTokens(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed looking up join tokens: %w", err)
	}

	result := make([]*entity.JoinToken, len(tokens))
	for i, t := range tokens {
		result[i] = t.ToEntity()
	}

	return result, nil
}

func (d *Datastore) UpdateJoinToken(ctx context.Context, joinTokenID uuid.UUID, used bool) (*entity.JoinToken, error) {
	pgID, err := uuidToPgType(joinTokenID)
	if err != nil {
		return nil, err
	}

	params := UpdateJoinTokenParams{
		ID:   pgID,
		Used: used,
	}

	jt, err := d.querier.UpdateJoinToken(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("failed updating join token with ID=%q, %w", joinTokenID, err)
	}

	return jt.ToEntity(), nil
}

func (d *Datastore) DeleteJoinToken(ctx context.Context, joinTokenID uuid.UUID) error {
	pgID, err := uuidToPgType(joinTokenID)
	if err != nil {
		return err
	}

	if err = d.querier.DeleteJoinToken(ctx, pgID); err != nil {
		return fmt.Errorf("failed deleting join token with ID=%q, %w", joinTokenID, err)
	}

	return nil
}

func (d *Datastore) FindJoinToken(ctx context.Context, token string) (*entity.JoinToken, error) {
	joinToken, err := d.querier.FindJoinToken(ctx, token)
	switch {
	case errors.Is(err, sql.ErrNoRows):
		return nil, nil
	case err != nil:
		return nil, fmt.Errorf("failed looking up join token: %w", err)
	}

	return joinToken.ToEntity(), nil
}

func (d *Datastore) CreateOrUpdateRelationship(ctx context.Context, req *entity.Relationship) (*entity.Relationship, error) {
	var relationship *Relationship
	var err error
	if req.ID.Valid {
		relationship, err = d.updateRelationship(ctx, req)
	} else {
		relationship, err = d.createRelationship(ctx, req)
	}
	if err != nil {
		return nil, err
	}

	response, err := relationship.ToEntity()
	if err != nil {
		return nil, fmt.Errorf("failed converting relationship model to entity: %w", err)
	}

	return response, nil
}

func (d *Datastore) FindRelationshipByID(ctx context.Context, relationshipID uuid.UUID) (*entity.Relationship, error) {
	pgID, err := uuidToPgType(relationshipID)
	if err != nil {
		return nil, err
	}

	relationship, err := d.querier.FindRelationshipByID(ctx, pgID)
	switch {
	case errors.Is(err, sql.ErrNoRows):
		return nil, nil
	case err != nil:
		return nil, fmt.Errorf("failed looking up relationship for ID=%q: %w", relationshipID, err)
	}

	response, err := relationship.ToEntity()
	if err != nil {
		return nil, fmt.Errorf("failed converting relationship model to entity: %w", err)
	}

	return response, nil
}

func (d *Datastore) FindRelationshipsByTrustDomainID(ctx context.Context, trustDomainID uuid.UUID) ([]*entity.Relationship, error) {
	pgID, err := uuidToPgType(trustDomainID)
	if err != nil {
		return nil, err
	}

	relationships, err := d.querier.FindRelationshipsByTrustDomainID(ctx, pgID)
	switch {
	case errors.Is(err, sql.ErrNoRows):
		return nil, nil
	case err != nil:
		return nil, fmt.Errorf("failed looking up relationships for TrustDomainID %q: %w", trustDomainID, err)
	}

	result := make([]*entity.Relationship, len(relationships))
	for i, m := range relationships {
		ent, err := m.ToEntity()
		if err != nil {
			return nil, fmt.Errorf("failed converting relationship model to entity: %w", err)
		}
		result[i] = ent
	}

	return result, nil
}

func (d *Datastore) ListRelationships(ctx context.Context) ([]*entity.Relationship, error) {
	relationships, err := d.querier.ListRelationships(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed looking up relationships: %w", err)
	}

	result := make([]*entity.Relationship, len(relationships))
	for i, m := range relationships {
		ent, err := m.ToEntity()
		if err != nil {
			return nil, fmt.Errorf("failed converting relationship model to entity: %w", err)
		}
		result[i] = ent
	}

	return result, nil
}

func (d *Datastore) DeleteRelationship(ctx context.Context, relationshipID uuid.UUID) error {
	pgID, err := uuidToPgType(relationshipID)
	if err != nil {
		return err
	}

	if err = d.querier.DeleteRelationship(ctx, pgID); err != nil {
		return fmt.Errorf("failed deleting relationship ID=%q: %w", relationshipID, err)
	}

	return nil
}

func (d *Datastore) createTrustDomain(ctx context.Context, req *entity.TrustDomain) (*TrustDomain, error) {
	params := CreateTrustDomainParams{
		Name: req.Name.String(),
	}
	if req.Description != "" {
		params.Description = sql.NullString{
			String: req.Description,
			Valid:  true,
		}
	}

	td, err := d.querier.CreateTrustDomain(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("failed creating new trust domain: %w", err)
	}
	return &td, nil
}

func (d *Datastore) updateTrustDomain(ctx context.Context, req *entity.TrustDomain) (*TrustDomain, error) {
	pgID, err := uuidToPgType(req.ID.UUID)
	if err != nil {
		return nil, err
	}

	params := UpdateTrustDomainParams{
		ID: pgID,
	}

	if req.Description != "" {
		params.Description = sql.NullString{
			String: req.Description,
			Valid:  true,
		}
	}

	td, err := d.querier.UpdateTrustDomain(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("failed updating trust domain: %w", err)
	}
	return &td, nil
}

func (d *Datastore) createBundle(ctx context.Context, req *entity.Bundle) (*Bundle, error) {
	pgTrustDomainID, err := uuidToPgType(req.TrustDomainID)
	if err != nil {
		return nil, err
	}
	params := CreateBundleParams{
		Data:               req.Data,
		Digest:             req.Digest,
		Signature:          req.Signature,
		SigningCertificate: req.SigningCertificate,
		TrustDomainID:      pgTrustDomainID,
	}

	bundle, err := d.querier.CreateBundle(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("failed creating new bundle: %w", err)
	}

	return &bundle, nil
}

func (d *Datastore) updateBundle(ctx context.Context, req *entity.Bundle) (*Bundle, error) {
	pgID, err := uuidToPgType(req.ID.UUID)
	if err != nil {
		return nil, err
	}
	params := UpdateBundleParams{
		ID:                 pgID,
		Data:               req.Data,
		Digest:             req.Digest,
		Signature:          req.Signature,
		SigningCertificate: req.SigningCertificate,
	}

	bundle, err := d.querier.UpdateBundle(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("failed updating bundle: %w", err)
	}

	return &bundle, nil
}

func (d *Datastore) createRelationship(ctx context.Context, req *entity.Relationship) (*Relationship, error) {
	pgTrustDomainAID, err := uuidToPgType(req.TrustDomainAID)
	if err != nil {
		return nil, err
	}

	pgTrustDomainBID, err := uuidToPgType(req.TrustDomainBID)
	if err != nil {
		return nil, err
	}

	params := CreateRelationshipParams{
		TrustDomainAID: pgTrustDomainAID,
		TrustDomainBID: pgTrustDomainBID,
	}

	relationship, err := d.querier.CreateRelationship(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("failed creating new relationship: %w", err)
	}

	return &relationship, nil
}

func (d *Datastore) updateRelationship(ctx context.Context, req *entity.Relationship) (*Relationship, error) {
	pgID, err := uuidToPgType(req.ID.UUID)
	if err != nil {
		return nil, err
	}

	params := UpdateRelationshipParams{
		ID:                  pgID,
		TrustDomainAConsent: ConsentStatus(req.TrustDomainAConsent),
		TrustDomainBConsent: ConsentStatus(req.TrustDomainBConsent),
	}

	relationship, err := d.querier.UpdateRelationship(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("failed updating relationship: %w", err)
	}

	return &relationship, nil
}
