package datastore

import (
	"context"
	"database/sql"
	"embed"
	"errors"
	"fmt"

	commondata "github.com/HewlettPackard/galadriel/pkg/common/datastore"
	"github.com/HewlettPackard/galadriel/pkg/gca/api"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	"github.com/google/uuid"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/stdlib"
	_ "github.com/lib/pq"
	"github.com/sirupsen/logrus"
)

// to be used for migration of data schema
//
//go:embed migrations/*.sql
var fs embed.FS

// Define interfaces to be used by the API handler

type Datastore interface {
	CreateTrustDomain(ctx context.Context, req *api.TrustDomain) (*api.TrustDomain, error)
	UpdateTrustDomain(ctx context.Context, req *TrustDomain) (*api.TrustDomain, error)
	DeleteTrustDomain(ctx context.Context, trustDomainID uuid.UUID) error
	FindTrustDomainByID(ctx context.Context, trustDomainID uuid.UUID) (*api.TrustDomain, error)
	FindTrustDomainByName(ctx context.Context, trustDomainName string) (*api.TrustDomain, error)
	ListTrustDomains(ctx context.Context) ([]*api.TrustDomain, error)
	CreateJoinToken(ctx context.Context, req *api.JoinToken) (*api.JoinToken, error)
	DeleteJoinToken(ctx context.Context, tokenID uuid.UUID) error
	FindJoinToken(ctx context.Context, token string) (*api.JoinToken, error)
	FindJoinTokenByID(ctx context.Context, tokenID uuid.UUID) (*api.JoinToken, error)
	FindJoinTokenByTrustDomainName(ctx context.Context, trustDomainName string) ([]*api.JoinToken, error)
	FindJoinTokensByTrustDomainID(ctx context.Context, trustDomainID uuid.UUID) ([]*api.JoinToken, error)
	ListJoinTokens(ctx context.Context) ([]*api.JoinToken, error)
	UpdateJoinToken(ctx context.Context, tokenID uuid.UUID, used bool) (*api.JoinToken, error)
}

// Define SQLStore that provides connectivity to the db, logger, and access to methods in querier

type SQLDatastore struct {
	logger  logrus.FieldLogger
	db      *sql.DB
	querier Querier
}

// Constants for Migration

const currentDBVersion = 1

const scheme = "postgresql"

// NewSQLDatastore creates  and datastore object that connects to the database (Postgres) and provides methods
// to interact with the datastore via querier
// The connection string can be:
//	url in the format of postgres://user:secret@localhost:port/db?sslmode=disable
//	DNS string user=postgres password=secret host=localhost port=5432 database=db sslmode=disable

func NewSQLDatastore(logger logrus.FieldLogger, cString string) (*SQLDatastore, error) {

	// validate connection string structure
	conf, err := pgx.ParseConfig(cString)
	if err != nil {
		return nil, fmt.Errorf("failed to parse GCA Postgres connection string: %w", err)
	}

	db := stdlib.OpenDB(*conf)
	// validate schema and perform migrations
	// When a new migration is created, this version should be updated in order to force
	// the migrations to run when starting up the app.
	sourceInstance, err := iofs.New(fs, "migrations")
	if err != nil {
		return nil, err
	}
	err = commondata.ValidateAndMigrateSchema(db, currentDBVersion, scheme, sourceInstance)
	if err != nil {
		return nil, fmt.Errorf("failed to validate or migrate GCA schema: %w", err)
	}

	return &SQLDatastore{
		logger:  logger,
		db:      db,
		querier: New(db),
	}, nil
}

// close db connection
func (d *SQLDatastore) Close() error {

	return d.db.Close()

}

// Implement all methods in SQLStore

// creates a new trust domain
func (d *SQLDatastore) CreateTrustDomain(ctx context.Context, req *api.TrustDomain) (*api.TrustDomain, error) {

	tdParams := CreateTrustDomainParams{
		Name: req.Name.String(),
	}
	if req.Description != "" {
		tdParams.Description = sql.NullString{
			String: req.Description,
			Valid:  true,
		}
	}

	td, err := d.querier.CreateTrustDomain(ctx, tdParams)
	if err != nil {
		return nil, fmt.Errorf("failed to create a new GCA trust domain: %w", err)
	}

	response, err := td.ToEntity()
	if err != nil {
		return nil, err
	}

	return response, nil
}

// updates a new trust domain
func (d *SQLDatastore) UpdateTrustDomain(ctx context.Context, req *api.TrustDomain) (*api.TrustDomain, error) {

	// convert uuid.UUID to pgtype.UUID
	pgID, err := toPGType(req.ID.UUID)
	if err != nil {
		return nil, err
	}

	tdParams := UpdateTrustDomainParams{
		ID:               pgID,
		OnboardingBundle: req.OnboardingBundle,
	}
	if req.Description != "" {
		tdParams.Description = sql.NullString{
			String: req.Description,
			Valid:  true,
		}
	}

	if !req.HarvesterSpiffeID.IsZero() {
		tdParams.HarvesterSpiffeID = sql.NullString{
			String: req.HarvesterSpiffeID.String(),
			Valid:  true,
		}
	}

	td, err := d.querier.UpdateTrustDomain(ctx, tdParams)
	if err != nil {
		return nil, fmt.Errorf("failed to update GCA trust domain: %w", err)
	}

	response, err := td.ToEntity()
	if err != nil {
		return nil, err
	}

	return response, nil
}

func (d *SQLDatastore) DeleteTrustDomain(ctx context.Context, trustDomainID uuid.UUID) error {

	// convert uuid.UUID to pgtype.UUID
	pgID, err := toPGType(trustDomainID)
	if err != nil {
		return err
	}

	err = d.querier.DeleteTrustDomain(ctx, pgID)
	if err != nil {
		return fmt.Errorf("failed deleting GCA trust domain: %w", err)
	}
	return nil
}

func (d *SQLDatastore) FindTrustDomainByID(ctx context.Context, trustDomainID uuid.UUID) (*api.TrustDomain, error) {

	// convert uuid.UUID to pgtype.UUID
	pgID, err := toPGType(trustDomainID)
	if err != nil {
		return nil, err
	}

	var td TrustDomain
	td, err = d.querier.FindTrustDomainByID(ctx, pgID)
	switch {
	case errors.Is(err, sql.ErrNoRows):
		return nil, nil
	case err != nil:
		return nil, fmt.Errorf("failed looking up GCA trust domain for ID=%q: %w", trustDomainID, err)
	}

	response, err := td.ToEntity()
	if err != nil {
		return nil, err
	}

	return response, nil
}
func (d *SQLDatastore) FindTrustDomainByName(ctx context.Context, trustDomainName string) (*api.TrustDomain, error) {

	// validate name
	if trustDomainName == "" {
		return nil, fmt.Errorf("failed finding GCA trust domain by name, invalid trust domain name")
	}

	td, err := d.querier.FindTrustDomainByName(ctx, trustDomainName)
	switch {
	case errors.Is(err, sql.ErrNoRows):
		return nil, nil
	case err != nil:
		return nil, fmt.Errorf("failed looking up GCA trust domain for ID=%q: %w", trustDomainName, err)
	}

	response, err := td.ToEntity()
	if err != nil {
		return nil, err
	}

	return response, nil
}
func (d *SQLDatastore) ListTrustDomains(ctx context.Context) ([]*api.TrustDomain, error) {

	trustdomains, err := d.querier.ListTrustDomains(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list GCA trust domains: %w", err)
	}

	result := make([]*api.TrustDomain, len(trustdomains))
	for i, td := range trustdomains {
		convtd, err := td.ToEntity()
		if err != nil {
			return nil, err
		}

		result[i] = convtd
	}

	return result, nil
}
func (d *SQLDatastore) CreateJoinToken(ctx context.Context, req *api.JoinToken) (*api.JoinToken, error) {

	pgID, err := toPGType(req.TrustDomainID)
	if err != nil {
		return nil, err
	}

	jtParams := CreateJoinTokenParams{
		TrustDomainID: pgID,
		Token:         req.Token,
		ExpiresAt:     req.ExpiresAt,
	}

	jt, err := d.querier.CreateJoinToken(ctx, jtParams)
	if err != nil {
		return nil, fmt.Errorf("failed creating a new GCA join token: %w", err)
	}

	response, err := jt.ToEntity()
	if err != nil {
		return nil, err
	}

	return response, nil
}
func (d *SQLDatastore) UpdateJoinToken(ctx context.Context, tokenID uuid.UUID, used bool) (*api.JoinToken, error) {

	// convert uuid.UUID to pgtype.UUID
	pgID, err := toPGType(tokenID)
	if err != nil {
		return nil, err
	}

	jtParams := UpdateJoinTokenParams{
		ID:   pgID,
		Used: used,
	}

	jt, err := d.querier.UpdateJoinToken(ctx, jtParams)
	if err != nil {
		return nil, fmt.Errorf("failed updating GCA join token: %w", err)
	}

	response, err := jt.ToEntity()
	if err != nil {
		return nil, err
	}

	return response, nil
}
func (d *SQLDatastore) FindJoinTokenByID(ctx context.Context, tokenID uuid.UUID) (*api.JoinToken, error) {

	// convert uuid.UUID to pgtype.UUID
	pgID, err := toPGType(tokenID)
	if err != nil {
		return nil, err
	}

	var jt JoinToken
	jt, err = d.querier.FindJoinTokenByID(ctx, pgID)
	switch {
	case errors.Is(err, sql.ErrNoRows):
		return nil, nil
	case err != nil:
		return nil, fmt.Errorf("failed looking up GCA join token for ID=%q: %w", tokenID, err)
	}

	response, err := jt.ToEntity()
	if err != nil {
		return nil, err
	}

	return response, nil
}
func (d *SQLDatastore) FindJoinTokenByTrustDomainName(ctx context.Context, trustDomainName string) ([]*api.JoinToken, error) {

	// validate name
	if trustDomainName == "" {
		return nil, fmt.Errorf("failed listing GCA Join tokens by Trust Domains Name, invalid trust domain name")
	}

	joinTokens, err := d.querier.FindJoinTokensByTrustDomainName(ctx, trustDomainName)
	switch {
	case errors.Is(err, sql.ErrNoRows):
		return nil, nil
	case err != nil:
		return nil, fmt.Errorf("failed looking up GCA join token for trustdomain name=%q: %w", trustDomainName, err)
	}

	result := make([]*api.JoinToken, len(joinTokens))
	for i, jt := range joinTokens {
		convljt, err := jt.ToEntity()
		if err != nil {
			return nil, err
		}
		result[i] = convljt
	}
	return result, nil
}

func (d *SQLDatastore) FindJoinToken(ctx context.Context, token string) (*api.JoinToken, error) {
	// convert uuid.UUID to pgtype.UUID
	// validate token
	if token == "" {
		return nil, fmt.Errorf("failed listing GCA join tokens by token, invalid token")
	}

	jt, err := d.querier.FindJoinToken(ctx, token)
	switch {
	case errors.Is(err, sql.ErrNoRows):
		return nil, nil
	case err != nil:
		return nil, fmt.Errorf("failed looking up GCA join token for token =%q: %w", token, err)
	}

	response, err := jt.ToEntity()
	if err != nil {
		return nil, err
	}

	return response, nil
}

func (d *SQLDatastore) FindJoinTokensByTrustDomainID(ctx context.Context, trustDomainID uuid.UUID) ([]*api.JoinToken, error) {
	// convert uuid.UUID to pgtype.UUID
	pgID, err := toPGType(trustDomainID)
	if err != nil {
		return nil, err
	}

	joinTokens, err := d.querier.FindJoinTokensByTrustDomainID(ctx, pgID)
	if err != nil {
		return nil, fmt.Errorf("failed to list GCA join tokens by trust domain ID: %w", err)
	}

	result := make([]*api.JoinToken, len(joinTokens))
	for i, jt := range joinTokens {
		convljt, err := jt.ToEntity()
		if err != nil {
			return nil, err
		}
		result[i] = convljt
	}
	return result, nil
}

func (d *SQLDatastore) ListJoinTokens(ctx context.Context) ([]*api.JoinToken, error) {

	joinTokens, err := d.querier.ListJoinTokens(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list GCA join tokens : %w", err)
	}

	result := make([]*api.JoinToken, len(joinTokens))
	for i, jt := range joinTokens {
		convljt, err := jt.ToEntity()
		if err != nil {
			return nil, err
		}
		result[i] = convljt
	}

	return result, nil
}

func (d *SQLDatastore) DeleteJoinToken(ctx context.Context, tokenID uuid.UUID) error {

	// convert uuid.UUID to pgtype.UUID
	pgID, err := toPGType(tokenID)
	if err != nil {
		return err
	}

	err = d.querier.DeleteJoinToken(ctx, pgID)
	if err != nil {
		return fmt.Errorf("failed to delete GCA join token: %w", err)
	}

	return nil
}
