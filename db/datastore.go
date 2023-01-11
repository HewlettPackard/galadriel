package db

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/HewlettPackard/galadriel/entity"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/stdlib"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"github.com/spiffe/go-spiffe/v2/spiffeid"
)

// Datastore is a Postgres database accessor that provides convenient methods
// to perform CRUD operations for Galadriel entities.
type Datastore struct {
	logger  logrus.FieldLogger
	db      *sql.DB
	querier Querier
}

// NewDatastore creates a new instance of a Datastore object that connects to a Postgres database
// parsing the connString.
// The connString can be a URL, e.g, "postgresql://host...", or a DSN, e.g., "host= user= password= dbname= port="
func NewDatastore(logger logrus.FieldLogger, connString string) (*Datastore, error) {
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
		logger:  logger,
		db:      db,
		querier: New(db),
	}, nil
}

func (d *Datastore) Close() error {
	return d.db.Close()
}

// CreateOrUpdateMember creates or updates the given Member in the underlying datastore, based on
// whether the given entity has an ID, in which case, it is updated.
func (d *Datastore) CreateOrUpdateMember(ctx context.Context, req *entity.Member) (*entity.Member, error) {
	if req.Status == "" {
		return nil, errors.New("member status is missing")
	}
	if req.TrustDomain.String() == "" {
		return nil, errors.New("member trust domain is missing")
	}

	var member *Member
	var err error
	if req.ID.Valid {
		member, err = d.updateMember(ctx, req)
	} else {
		member, err = d.createMember(ctx, req)
	}
	if err != nil {
		return nil, err
	}

	response, err := member.ToEntity()
	if err != nil {
		return nil, fmt.Errorf("failed converting member model to entity: %w", err)
	}

	return response, nil
}

func (d *Datastore) createMember(ctx context.Context, req *entity.Member) (*Member, error) {
	var status Status
	err := status.Scan(req.Status.String())
	if err != nil {
		return nil, fmt.Errorf("failed parsing status: %w", err)
	}

	params := CreateMemberParams{
		TrustDomain: req.TrustDomain.String(),
		Status:      status,
	}

	member, err := d.querier.CreateMember(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("failed creating new member: %w", err)
	}
	return &member, nil
}

func (d *Datastore) updateMember(ctx context.Context, req *entity.Member) (*Member, error) {
	pgID, err := uuidToPgType(req.ID.UUID)
	if err != nil {
		return nil, err
	}

	var status Status
	err = status.Scan(req.Status.String())
	if err != nil {
		return nil, fmt.Errorf("failed parsing status: %w", err)
	}

	params := UpdateMemberParams{
		ID:          pgID,
		TrustDomain: req.TrustDomain.String(),
		Status:      status,
	}

	member, err := d.querier.UpdateMember(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("failed updating member: %w", err)
	}
	return &member, nil
}

func (d *Datastore) DeleteMember(ctx context.Context, memberID uuid.UUID) error {
	pgID, err := uuidToPgType(memberID)
	if err != nil {
		return err
	}

	if err = d.querier.DeleteMember(ctx, pgID); err != nil {
		return fmt.Errorf("failed deleting member with ID=%q: %w", memberID, err)
	}

	return nil
}

func (d *Datastore) ListMembers(ctx context.Context) ([]*entity.Member, error) {
	members, err := d.querier.ListMembers(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed getting member list: %w", err)
	}

	result := make([]*entity.Member, len(members))
	for i, m := range members {
		r, err := m.ToEntity()
		if err != nil {
			return nil, fmt.Errorf("failed converting model member to entity: %v", err)
		}
		result[i] = r
	}

	return result, nil
}

func (d *Datastore) FindMemberByID(ctx context.Context, memberID uuid.UUID) (*entity.Member, error) {
	pgID, err := uuidToPgType(memberID)
	if err != nil {
		return nil, err
	}

	m, err := d.querier.FindMemberByID(ctx, pgID)
	switch {
	case errors.Is(err, sql.ErrNoRows):
		return nil, nil
	case err != nil:
		return nil, fmt.Errorf("failed looking up member for ID=%q: %w", memberID, err)
	}

	r, err := m.ToEntity()
	if err != nil {
		return nil, fmt.Errorf("failed converting model member to entity: %w", err)
	}

	return r, nil
}

func (d *Datastore) FindMemberByTrustDomain(ctx context.Context, trustDomain spiffeid.TrustDomain) (*entity.Member, error) {
	m, err := d.querier.FindMemberByTrustDomain(ctx, trustDomain.String())
	switch {
	case errors.Is(err, sql.ErrNoRows):
		return nil, nil
	case err != nil:
		return nil, fmt.Errorf("failed looking up member for Trust Domain=%q: %w", trustDomain, err)
	}

	r, err := m.ToEntity()
	if err != nil {
		return nil, fmt.Errorf("failed converting model member to entity: %w", err)
	}

	return r, nil
}

func (d *Datastore) CreateOrUpdateFederationGroup(ctx context.Context, req *entity.FederationGroup) (*entity.FederationGroup, error) {
	if req.Name == "" {
		return nil, errors.New("federation group name is missing")
	}
	if req.Status == "" {
		return nil, errors.New("federation group status is missing")
	}

	var fg *FederationGroup
	var err error
	if req.ID.Valid {
		fg, err = d.updateFederationGroup(ctx, req)
	} else {
		fg, err = d.createFederationGroup(ctx, req)
	}
	if err != nil {
		return nil, err
	}

	response, err := fg.ToEntity()
	if err != nil {
		return nil, fmt.Errorf("failed converting federation group model to entity: %w", err)
	}

	return response, nil
}

func (d *Datastore) createFederationGroup(ctx context.Context, req *entity.FederationGroup) (*FederationGroup, error) {
	var status Status
	err := status.Scan(req.Status.String())
	if err != nil {
		return nil, fmt.Errorf("failed parsing status: %w", err)
	}

	params := CreateFederationGroupParams{
		Name: req.Name,
		Description: sql.NullString{
			String: req.Description,
			Valid:  true,
		},
		Status: status,
	}
	fg, err := d.querier.CreateFederationGroup(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("failed creating new federation group: %w", err)
	}

	return &fg, nil
}

func (d *Datastore) updateFederationGroup(ctx context.Context, req *entity.FederationGroup) (*FederationGroup, error) {
	pgID, err := uuidToPgType(req.ID.UUID)
	if err != nil {
		return nil, err
	}

	var status Status
	err = status.Scan(req.Status.String())
	if err != nil {
		return nil, fmt.Errorf("failed parsing status: %w", err)
	}

	params := UpdateFederationGroupParams{
		ID:   pgID,
		Name: req.Name,
		Description: sql.NullString{
			String: req.Description,
			Valid:  true,
		},
		Status: status,
	}

	fg, err := d.querier.UpdateFederationGroup(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("failed creating new federation group: %w", err)
	}

	return &fg, nil
}

func (d *Datastore) FindFederationGroupByID(ctx context.Context, federatioGroupID uuid.UUID) (*entity.FederationGroup, error) {
	pgID, err := uuidToPgType(federatioGroupID)
	if err != nil {
		return nil, err
	}

	m, err := d.querier.FindFederationGroupByID(ctx, pgID)
	switch {
	case errors.Is(err, sql.ErrNoRows):
		return nil, nil
	case err != nil:
		return nil, fmt.Errorf("failed looking up federation group with ID=%q: %w", federatioGroupID, err)
	}

	r, err := m.ToEntity()
	if err != nil {
		return nil, fmt.Errorf("failed converting model federation group to entity: %w", err)
	}

	return r, nil
}

func (d *Datastore) DeleteFederationGroup(ctx context.Context, federationGroupID uuid.UUID) error {
	pgID, err := uuidToPgType(federationGroupID)
	if err != nil {
		return err
	}

	if err = d.querier.DeleteFederationGroup(ctx, pgID); err != nil {
		return fmt.Errorf("failed deleting federation group with ID=%q: %w", federationGroupID, err)
	}

	return nil
}

func (d *Datastore) ListFederationGroups(ctx context.Context) ([]*entity.FederationGroup, error) {
	fgs, err := d.querier.ListFederationGroups(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed getting federation groups list: %w", err)
	}

	result := make([]*entity.FederationGroup, len(fgs))
	for i, m := range fgs {
		r, err := m.ToEntity()
		if err != nil {
			return nil, fmt.Errorf("failed converting model federation group to entity: %w", err)
		}
		result[i] = r
	}

	return result, nil
}

func (d *Datastore) CreateOrUpdateBundle(ctx context.Context, req *entity.Bundle) (*entity.Bundle, error) {
	if !req.MemberID.Valid {
		return nil, errors.New("bundle member ID cannot be empty")
	}

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
		return nil, fmt.Errorf("failed converting member model to entity: %w", err)
	}

	return response, nil
}

func (d *Datastore) createBundle(ctx context.Context, req *entity.Bundle) (*Bundle, error) {
	pgMemberID, err := uuidToPgType(req.MemberID.UUID)
	if err != nil {
		return nil, err
	}
	params := CreateBundleParams{
		RawBundle:    req.RawBundle,
		Digest:       req.Digest,
		SignedBundle: req.SignedBundle,
		TlogID:       req.TlogID,
		SvidPem: sql.NullString{
			String: req.SvidPem,
			Valid:  true,
		},
		MemberID: pgMemberID,
	}

	bundle, err := d.querier.CreateBundle(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("failed creating new bundle: %w", err)
	}

	return &bundle, err
}

func (d *Datastore) updateBundle(ctx context.Context, req *entity.Bundle) (*Bundle, error) {
	pgID, err := uuidToPgType(req.ID.UUID)
	if err != nil {
		return nil, err
	}
	params := UpdateBundleParams{
		ID:           pgID,
		RawBundle:    req.RawBundle,
		Digest:       req.Digest,
		SignedBundle: req.SignedBundle,
		TlogID:       req.TlogID,
		SvidPem: sql.NullString{
			String: req.SvidPem,
			Valid:  true,
		},
	}

	bundle, err := d.querier.UpdateBundle(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("failed updating bundle: %w", err)
	}

	return &bundle, nil
}

func (d *Datastore) FindBundleByID(ctx context.Context, bundleID uuid.UUID) (*entity.Bundle, error) {
	pgID, err := uuidToPgType(bundleID)
	if err != nil {
		return nil, err
	}

	m, err := d.querier.FindBundleByID(ctx, pgID)
	switch {
	case errors.Is(err, sql.ErrNoRows):
		return nil, nil
	case err != nil:
		return nil, fmt.Errorf("failed looking up bundle with ID=%q: %w", bundleID, err)
	}

	r, err := m.ToEntity()
	if err != nil {
		return nil, fmt.Errorf("failed converting model bundle to entity: %w", err)
	}

	return r, nil
}

func (d *Datastore) FindBundleByMemberID(ctx context.Context, memberID uuid.UUID) (*entity.Bundle, error) {
	pgID, err := uuidToPgType(memberID)
	if err != nil {
		return nil, err
	}

	m, err := d.querier.FindBundleByMemberID(ctx, pgID)
	switch {
	case errors.Is(err, sql.ErrNoRows):
		return nil, nil
	case err != nil:
		return nil, fmt.Errorf("failed looking up bundle for ID=%q: %w", memberID, err)
	}

	r, err := m.ToEntity()
	if err != nil {
		return nil, fmt.Errorf("failed converting model bundle to entity: %w", err)
	}

	return r, nil
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
	pgID, err := uuidToPgType(req.MemberID.UUID)
	if err != nil {
		return nil, err
	}

	params := CreateJoinTokenParams{
		Token:    req.Token,
		Expiry:   req.Expiry,
		MemberID: pgID,
	}
	token, err := d.querier.CreateJoinToken(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("failed creating join token: %w", err)
	}

	return token.ToEntity(), nil
}

func (d *Datastore) FindJoinTokensByID(ctx context.Context, joinTokenID uuid.UUID) (*entity.JoinToken, error) {
	pgID, err := uuidToPgType(joinTokenID)
	if err != nil {
		return nil, err
	}

	t, err := d.querier.FindJoinTokenByID(ctx, pgID)
	switch {
	case errors.Is(err, sql.ErrNoRows):
		return nil, nil
	case err != nil:
		return nil, fmt.Errorf("failed looking up join token with ID=%q: %w", joinTokenID, err)
	}

	return t.ToEntity(), nil
}

func (d *Datastore) FindJoinTokensByMemberID(ctx context.Context, memberID uuid.UUID) ([]*entity.JoinToken, error) {
	pgID, err := uuidToPgType(memberID)
	if err != nil {
		return nil, err
	}

	tokens, err := d.querier.FindJoinTokensByMemberID(ctx, pgID)
	switch {
	case errors.Is(err, sql.ErrNoRows):
		return nil, nil
	case err != nil:
		return nil, fmt.Errorf("failed looking up join token for Member ID=%q: %w", memberID, err)
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
		ID: pgID,
		Used: sql.NullBool{
			Bool:  used,
			Valid: true,
		},
	}

	t, err := d.querier.UpdateJoinToken(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("failed updating join token with ID=%q, %w", joinTokenID, err)
	}

	return t.ToEntity(), nil
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
	t, err := d.querier.FindJoinToken(ctx, token)
	switch {
	case errors.Is(err, sql.ErrNoRows):
		return nil, nil
	case err != nil:
		return nil, fmt.Errorf("failed looking up join token %q: %w", token, err)
	}

	return t.ToEntity(), nil
}

func (d *Datastore) CreateOrUpdateMembership(ctx context.Context, req *entity.Membership) (*entity.Membership, error) {
	if req.Status == "" {
		return nil, errors.New("Membership status is missing")
	}

	var membership *Membership
	var err error
	if req.ID.Valid {
		membership, err = d.updateMembership(ctx, req)
	} else {
		membership, err = d.createMembership(ctx, req)
	}
	if err != nil {
		return nil, err
	}

	response, err := membership.ToEntity()
	if err != nil {
		return nil, fmt.Errorf("failed converting membership model to entity: %w", err)
	}

	return response, nil
}

func (d *Datastore) createMembership(ctx context.Context, req *entity.Membership) (*Membership, error) {
	if !req.MemberID.Valid {
		return nil, errors.New("Member ID is missing")
	}
	pgMemberID, err := uuidToPgType(req.MemberID.UUID)
	if err != nil {
		return nil, err
	}

	if !req.FederationGroupID.Valid {
		return nil, errors.New("FederationGroup ID is missing")
	}
	pgFederationGroupID, err := uuidToPgType(req.FederationGroupID.UUID)
	if err != nil {
		return nil, err
	}

	var status Status
	err = status.Scan(req.Status.String())
	if err != nil {
		return nil, fmt.Errorf("failed parsing status: %w", err)
	}

	params := CreateMembershipParams{
		MemberID:          pgMemberID,
		FederationGroupID: pgFederationGroupID,
		Status:            status,
	}

	membership, err := d.querier.CreateMembership(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("failed creating new membership: %w", err)
	}

	return &membership, nil
}

func (d *Datastore) updateMembership(ctx context.Context, req *entity.Membership) (*Membership, error) {
	pgID, err := uuidToPgType(req.ID.UUID)
	if err != nil {
		return nil, err
	}

	var status Status
	err = status.Scan(req.Status.String())
	if err != nil {
		return nil, fmt.Errorf("failed parsing status: %w", err)
	}

	params := UpdateMembershipParams{
		ID:     pgID,
		Status: status,
	}

	membership, err := d.querier.UpdateMembership(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("failed updating membership: %w", err)
	}

	return &membership, nil
}

func (d *Datastore) FindMembershipByID(ctx context.Context, membershipID uuid.UUID) (*entity.Membership, error) {
	pgID, err := uuidToPgType(membershipID)
	if err != nil {
		return nil, err
	}

	m, err := d.querier.FindMembershipByID(ctx, pgID)
	switch {
	case errors.Is(err, sql.ErrNoRows):
		return nil, nil
	case err != nil:
		return nil, fmt.Errorf("failed looking up membership for ID=%q: %w", membershipID, err)
	}

	response, err := m.ToEntity()
	if err != nil {
		return nil, fmt.Errorf("failed converting membership model to entity: %w", err)
	}

	return response, nil
}

func (d *Datastore) FindMembershipsByMemberID(ctx context.Context, memberID uuid.UUID) ([]*entity.Membership, error) {
	pgID, err := uuidToPgType(memberID)
	if err != nil {
		return nil, err
	}

	memberships, err := d.querier.FindMembershipsByMemberID(ctx, pgID)
	switch {
	case errors.Is(err, sql.ErrNoRows):
		return nil, nil
	case err != nil:
		return nil, fmt.Errorf("failed looking up memberships for MemberID %q: %w", memberID, err)
	}

	result := make([]*entity.Membership, len(memberships))
	for i, m := range memberships {
		ent, err := m.ToEntity()
		if err != nil {
			return nil, fmt.Errorf("failed converting membership model to entity: %w", err)
		}
		result[i] = ent
	}

	return result, nil
}

func (d *Datastore) ListMemberships(ctx context.Context) ([]*entity.Membership, error) {
	memberships, err := d.querier.ListMemberships(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed looking up memberships: %w", err)
	}

	result := make([]*entity.Membership, len(memberships))
	for i, m := range memberships {
		ent, err := m.ToEntity()
		if err != nil {
			return nil, fmt.Errorf("failed converting membership model to entity: %w", err)
		}
		result[i] = ent
	}

	return result, nil
}

func (d *Datastore) DeleteMembership(ctx context.Context, id uuid.UUID) error {
	pgID, err := uuidToPgType(id)
	if err != nil {
		return err
	}

	if err = d.querier.DeleteMembership(ctx, pgID); err != nil {
		return fmt.Errorf("failed deleting membership ID=%q: %w", id, err)
	}

	return nil
}

func (d *Datastore) CreateOrUpdateHarvester(ctx context.Context, req *entity.Harvester) (*entity.Harvester, error) {
	var harvester *Harvester
	var err error
	if req.ID.Valid {
		harvester, err = d.updateHarvester(ctx, req)
	} else {
		harvester, err = d.createHarvester(ctx, req)
	}
	if err != nil {
		return nil, err
	}

	return harvester.ToEntity(), nil
}

func (d *Datastore) updateHarvester(ctx context.Context, req *entity.Harvester) (*Harvester, error) {
	pgID, err := uuidToPgType(req.ID.UUID)
	if err != nil {
		return nil, err
	}

	params := UpdateHarvesterParams{
		ID: pgID,
		IsLeader: sql.NullBool{
			Bool:  req.IsLeader,
			Valid: true,
		},
		LeaderUntil: req.LeaderUntil,
	}

	harvester, err := d.querier.UpdateHarvester(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("failed updating harvester: %w", err)
	}

	return &harvester, nil
}

func (d *Datastore) createHarvester(ctx context.Context, req *entity.Harvester) (*Harvester, error) {
	if !req.MemberID.Valid {
		return nil, errors.New("harvester member ID cannot be empty")
	}
	pgMemberID, err := uuidToPgType(req.MemberID.UUID)
	if err != nil {
		return nil, err
	}

	params := CreateHarvesterParams{
		MemberID: pgMemberID,
		IsLeader: sql.NullBool{
			Bool:  req.IsLeader,
			Valid: true,
		},
		LeaderUntil: req.LeaderUntil,
	}

	harvester, err := d.querier.CreateHarvester(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("failed creating new harvester: %w", err)
	}

	return &harvester, nil
}

func (d *Datastore) FindHarvesterByID(ctx context.Context, harvesterID uuid.UUID) (*entity.Harvester, error) {
	pgID, err := uuidToPgType(harvesterID)
	if err != nil {
		return nil, err
	}

	h, err := d.querier.FindHarvesterByID(ctx, pgID)
	switch {
	case errors.Is(err, sql.ErrNoRows):
		return nil, nil
	case err != nil:
		return nil, fmt.Errorf("failed looking up harvester ID=%q: %w", harvesterID, err)
	}

	return h.ToEntity(), nil
}

func (d *Datastore) FindHarvestersByMemberID(ctx context.Context, memberID uuid.UUID) ([]*entity.Harvester, error) {
	pgMemberID, err := uuidToPgType(memberID)
	if err != nil {
		return nil, err
	}

	harvesters, err := d.querier.FindHarvestersByMemberID(ctx, pgMemberID)
	switch {
	case errors.Is(err, sql.ErrNoRows):
		return nil, nil
	case err != nil:
		return nil, fmt.Errorf("failed looking up harvesters with MemberID=%q: %w", memberID, err)
	}

	result := make([]*entity.Harvester, len(harvesters))
	for i, h := range harvesters {
		result[i] = h.ToEntity()
	}

	return result, nil
}

func (d *Datastore) ListHarvesters(ctx context.Context) ([]*entity.Harvester, error) {
	harvesters, err := d.querier.ListHarvesters(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed looking up harvesters: %w", err)
	}

	result := make([]*entity.Harvester, len(harvesters))
	for i, h := range harvesters {
		result[i] = h.ToEntity()
	}

	return result, nil
}

func (d *Datastore) DeleteHarvester(ctx context.Context, harvesterID uuid.UUID) error {
	pgID, err := uuidToPgType(harvesterID)
	if err != nil {
		return err
	}

	if err = d.querier.DeleteHarvester(ctx, pgID); err != nil {
		return fmt.Errorf("failed deleting harvester ID=%q: %w", harvesterID, err)
	}

	return nil
}
