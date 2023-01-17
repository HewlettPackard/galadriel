package db

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/HewlettPackard/galadriel/entity"
	"github.com/google/uuid"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/ory/dockertest/v3"
	"github.com/ory/dockertest/v3/docker"
	"github.com/sirupsen/logrus"
	"github.com/spiffe/go-spiffe/v2/spiffeid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	postgresImage = "15-alpine"
	user          = "test_user"
	password      = "test_password"
	dbname        = "test_db"
)

var (
	td1     = spiffeid.RequireTrustDomainFromString("foo.test")
	td2     = spiffeid.RequireTrustDomainFromString("bar.test")
	otherTd = spiffeid.RequireTrustDomainFromString("other.test")
)

func TestCRUDFederationGroup(t *testing.T) {
	t.Parallel()
	datastore, err := setupDatastore(t)
	require.NoError(t, err, "Failed to setup the datastore")

	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)

	// Create Federation Group
	req1 := &entity.FederationGroup{
		Name:        "fg1-test",
		Description: "test-federation-group",
		Status:      entity.StatusActive,
	}
	fg1, err := datastore.CreateOrUpdateFederationGroup(ctx, req1)
	require.NoError(t, err)
	require.NotNil(t, fg1.ID)
	assert.Equal(t, req1.Name, fg1.Name)
	assert.Equal(t, req1.Description, fg1.Description)
	assert.Equal(t, req1.Status, fg1.Status)

	// Look up federation group stored in DB and compare
	stored1, err := datastore.FindFederationGroupByID(ctx, fg1.ID.UUID)
	require.NoError(t, err)
	assert.Equal(t, fg1, stored1)

	// Create second Federation Group
	req2 := &entity.FederationGroup{
		Name:        "fg2-test",
		Description: "test-federation-group-2",
		Status:      entity.StatusPending,
	}
	fg2, err := datastore.CreateOrUpdateFederationGroup(ctx, req2)
	require.NoError(t, err)
	require.NotNil(t, fg2.ID)
	assert.Equal(t, req2.Name, fg2.Name)
	assert.Equal(t, req2.Description, fg2.Description)
	assert.Equal(t, req2.Status, fg2.Status)

	// Look up federation group stored in DB and compare
	stored2, err := datastore.FindFederationGroupByID(ctx, fg2.ID.UUID)
	require.NoError(t, err)
	assert.Equal(t, fg2, stored2)

	// Update first federation group
	fg1.Status = entity.StatusDisabled
	fg1.Name = "other-name"
	fg1.Description = "other-description"
	updated1, err := datastore.CreateOrUpdateFederationGroup(ctx, fg1)
	require.NoError(t, err)
	require.NotNil(t, updated1)
	assert.Equal(t, entity.StatusDisabled, updated1.Status)
	assert.Equal(t, fg1.Name, updated1.Name)
	assert.Equal(t, fg1.Description, updated1.Description)

	// Look up federation group stored in DB and compare
	stored1, err = datastore.FindFederationGroupByID(ctx, fg1.ID.UUID)
	require.NoError(t, err)
	assert.Equal(t, updated1, stored1)

	// List Federation Groups
	fedGroups, err := datastore.ListFederationGroups(ctx)
	require.NoError(t, err)
	assert.Equal(t, 2, len(fedGroups))
	require.Contains(t, fedGroups, stored1)
	require.Contains(t, fedGroups, stored2)

	// Delete federation group
	err = datastore.DeleteFederationGroup(ctx, fg1.ID.UUID)
	require.NoError(t, err)

	// Check that the deleted member is no longer in the DB
	found, err := datastore.FindFederationGroupByID(ctx, fg1.ID.UUID)
	require.NoError(t, err)
	require.Nil(t, found)

	// and that one federation group remains
	fedGroups, err = datastore.ListFederationGroups(ctx)
	require.NoError(t, err)
	assert.Equal(t, 1, len(fedGroups))
	assert.Equal(t, fg2, fedGroups[0])
}

func TestFederationGroupUniqueNameConstraint(t *testing.T) {
	t.Parallel()
	datastore, err := setupDatastore(t)
	require.NoError(t, err, "Failed to setup the datastore")

	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)

	// Create Federation Group
	fg1 := &entity.FederationGroup{
		Name:        "fed-group-name",
		Description: "test-federation-group",
		Status:      entity.StatusActive,
	}
	fg1, err = datastore.CreateOrUpdateFederationGroup(ctx, fg1)
	require.NoError(t, err)
	require.NotNil(t, fg1.ID)

	//// Create Federation Group
	fg2 := &entity.FederationGroup{
		Name:        "fed-group-name",
		Description: "other-fed-group",
		Status:      entity.StatusPending,
	}
	fg2, err = datastore.CreateOrUpdateFederationGroup(ctx, fg2)
	require.Error(t, err)
	require.Nil(t, fg2)

	wrappedErr := errors.Unwrap(err)
	errCode := wrappedErr.(*pgconn.PgError).SQLState()
	assert.Equal(t, pgerrcode.UniqueViolation, errCode, "Unique constraint violation was expected")
}

func TestCRUDMember(t *testing.T) {
	t.Parallel()
	datastore, err := setupDatastore(t)
	require.NoError(t, err, "Failed to setup the datastore")

	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)

	// Create Member
	req1 := &entity.Member{
		TrustDomain: td1,
		Status:      entity.StatusActive,
	}
	m1, err := datastore.CreateOrUpdateMember(ctx, req1)
	require.NoError(t, err)
	require.NotNil(t, m1.ID)
	assert.Equal(t, req1.TrustDomain, m1.TrustDomain)
	assert.Equal(t, req1.Status, m1.Status)
	require.NotNil(t, m1.CreatedAt)
	require.NotNil(t, m1.UpdatedAt)

	// Look up member stored in DB and compare
	stored, err := datastore.FindMemberByID(ctx, m1.ID.UUID)
	require.NoError(t, err)
	assert.Equal(t, m1, stored)

	// Create second member
	req2 := &entity.Member{
		TrustDomain: td2,
		Status:      entity.StatusPending,
	}
	m2, err := datastore.CreateOrUpdateMember(ctx, req2)
	require.NoError(t, err)
	require.NotNil(t, m2.ID)
	assert.Equal(t, req2.TrustDomain, m2.TrustDomain)
	assert.Equal(t, req2.Status, m2.Status)
	require.NotNil(t, m2.CreatedAt)
	require.NotNil(t, m2.UpdatedAt)

	// Look up member stored in DB and compare
	stored, err = datastore.FindMemberByID(ctx, m2.ID.UUID)
	require.NoError(t, err)
	assert.Equal(t, m2, stored)

	// Update First Member
	m1.Status = entity.StatusDisabled
	updated1, err := datastore.CreateOrUpdateMember(ctx, m1)
	require.NoError(t, err)
	require.NotNil(t, updated1)
	assert.Equal(t, entity.StatusDisabled, updated1.Status)

	// Look up member stored in DB and compare
	stored, err = datastore.FindMemberByID(ctx, m1.ID.UUID)
	require.NoError(t, err)
	assert.Equal(t, updated1, stored)

	// Find members by Trust Domain
	found1, err := datastore.FindMemberByTrustDomain(ctx, m1.TrustDomain)
	require.NoError(t, err)
	require.NotNil(t, found1)
	assert.Equal(t, updated1, found1)

	found2, err := datastore.FindMemberByTrustDomain(ctx, m2.TrustDomain)
	require.NoError(t, err)
	require.NotNil(t, found2)
	assert.Equal(t, m2, found2)

	// Look up non-existent member
	found, err := datastore.FindMemberByTrustDomain(ctx, otherTd)
	require.NoError(t, err)
	require.Nil(t, found)

	// List members
	members, err := datastore.ListMembers(ctx)
	require.NoError(t, err)
	assert.Equal(t, 2, len(members))
	require.Contains(t, members, found1)
	require.Contains(t, members, found2)

	// Delete member
	err = datastore.DeleteMember(ctx, m1.ID.UUID)
	require.NoError(t, err)

	// Check that the deleted member is no longer in the DB
	found, err = datastore.FindMemberByID(ctx, m1.ID.UUID)
	require.NoError(t, err)
	require.Nil(t, found)

	// and that one member remains
	members, err = datastore.ListMembers(ctx)
	require.NoError(t, err)
	assert.Equal(t, 1, len(members))
	assert.Equal(t, m2, members[0])

	// Delete the other member
	err = datastore.DeleteMember(ctx, m2.ID.UUID)
	require.NoError(t, err)

	// Check that all members were deleted
	members, err = datastore.ListMembers(ctx)
	require.NoError(t, err)
	assert.Equal(t, 0, len(members))
}

func TestMemberUniqueTrustDomainConstraint(t *testing.T) {
	t.Parallel()
	datastore, err := setupDatastore(t)
	require.NoError(t, err, "Failed to setup the datastore")

	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)

	m1 := &entity.Member{
		TrustDomain: td1,
		Status:      entity.StatusActive,
	}
	_, err = datastore.CreateOrUpdateMember(ctx, m1)
	require.NoError(t, err)

	// second member with same trust domain
	m2 := &entity.Member{
		TrustDomain: td1,
		Status:      entity.StatusPending,
	}
	_, err = datastore.CreateOrUpdateMember(ctx, m2)
	require.Error(t, err)

	wrappedErr := errors.Unwrap(err)
	errCode := wrappedErr.(*pgconn.PgError).SQLState()
	assert.Equal(t, pgerrcode.UniqueViolation, errCode, "Unique constraint violation was expected")
}

func TestCRUDMembership(t *testing.T) {
	t.Parallel()
	datastore, err := setupDatastore(t)
	require.NoError(t, err, "Failed to setup the datastore")

	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)

	// Setup
	// Create Federation Groups
	fg1 := &entity.FederationGroup{
		Name:        "fg1-test",
		Description: "test-federation-group",
		Status:      entity.StatusActive,
	}
	fg1, err = datastore.CreateOrUpdateFederationGroup(ctx, fg1)
	require.NoError(t, err)

	fg2 := &entity.FederationGroup{
		Name:        "fg2-test",
		Description: "test-federation-group-2",
		Status:      entity.StatusActive,
	}
	fg2, err = datastore.CreateOrUpdateFederationGroup(ctx, fg2)
	require.NoError(t, err)

	// Create Members
	m1 := &entity.Member{
		TrustDomain: td1,
		Status:      entity.StatusActive,
	}
	m1, err = datastore.CreateOrUpdateMember(ctx, m1)
	require.NoError(t, err)

	m2 := &entity.Member{
		TrustDomain: td2,
		Status:      entity.StatusActive,
	}
	m2, err = datastore.CreateOrUpdateMember(ctx, m2)
	require.NoError(t, err)

	// Create membership Member-1 -> FederationGroup-1
	req1 := &entity.Membership{
		MemberID:          m1.ID,
		FederationGroupID: fg1.ID,
		Status:            entity.StatusPending,
	}
	membership1, err := datastore.CreateOrUpdateMembership(ctx, req1)
	require.NoError(t, err)
	require.NotNil(t, membership1.ID)
	assert.Equal(t, req1.MemberID, membership1.MemberID)
	assert.Equal(t, req1.FederationGroupID, membership1.FederationGroupID)

	// Look up membership in DB and compare
	stored, err := datastore.FindMembershipByID(ctx, membership1.ID.UUID)
	require.NoError(t, err)
	assert.Equal(t, membership1, stored)

	// Create membership Member-1 -> FederationGroup-2
	req2 := &entity.Membership{
		MemberID:          m1.ID,
		FederationGroupID: fg2.ID,
		Status:            entity.StatusPending,
	}
	membership2, err := datastore.CreateOrUpdateMembership(ctx, req2)
	require.NoError(t, err)
	require.NotNil(t, membership2.ID)
	assert.Equal(t, req2.MemberID, membership2.MemberID)
	assert.Equal(t, req2.FederationGroupID, membership2.FederationGroupID)

	// Look up membership in DB and compare
	stored, err = datastore.FindMembershipByID(ctx, membership2.ID.UUID)
	require.NoError(t, err)
	assert.Equal(t, membership2, stored)

	// Create membership Member-2 -> FederationGroup-2
	req3 := &entity.Membership{
		MemberID:          m2.ID,
		FederationGroupID: fg2.ID,
		Status:            entity.StatusPending,
	}
	membership3, err := datastore.CreateOrUpdateMembership(ctx, req3)
	require.NoError(t, err)
	require.NotNil(t, membership3.ID)
	assert.Equal(t, req3.MemberID, membership3.MemberID)
	assert.Equal(t, req3.FederationGroupID, membership3.FederationGroupID)

	// Look up membership in DB and compare
	stored, err = datastore.FindMembershipByID(ctx, membership3.ID.UUID)
	require.NoError(t, err)
	assert.Equal(t, membership3, stored)

	// Find memberships by MemberID
	memberships, err := datastore.FindMembershipsByMemberID(ctx, m1.ID.UUID)
	require.NoError(t, err)
	assert.Equal(t, 2, len(memberships))
	memberships[0].MemberID = m1.ID
	memberships[1].MemberID = m1.ID

	memberships, err = datastore.FindMembershipsByMemberID(ctx, m2.ID.UUID)
	require.NoError(t, err)
	assert.Equal(t, 1, len(memberships))
	memberships[0].MemberID = m2.ID

	// List all memberships
	memberships, err = datastore.ListMemberships(ctx)
	require.NoError(t, err)
	assert.Equal(t, 3, len(memberships))

	// Update membership
	membership1.Status = entity.StatusActive
	updated, err := datastore.CreateOrUpdateMembership(ctx, membership1)
	require.NoError(t, err)
	assert.Equal(t, entity.StatusActive, updated.Status)

	// Look up membership in DB and compare
	stored, err = datastore.FindMembershipByID(ctx, membership1.ID.UUID)
	require.NoError(t, err)
	assert.Equal(t, updated, stored)

	// Delete memberships
	err = datastore.DeleteMembership(ctx, membership1.ID.UUID)
	require.NoError(t, err)
	stored, err = datastore.FindMembershipByID(ctx, membership1.ID.UUID)
	require.NoError(t, err)
	assert.Nil(t, stored)

	err = datastore.DeleteMembership(ctx, membership2.ID.UUID)
	require.NoError(t, err)
	stored, err = datastore.FindMembershipByID(ctx, membership2.ID.UUID)
	require.NoError(t, err)
	assert.Nil(t, stored)

	err = datastore.DeleteMembership(ctx, membership3.ID.UUID)
	require.NoError(t, err)
	stored, err = datastore.FindMembershipByID(ctx, membership3.ID.UUID)
	require.NoError(t, err)
	assert.Nil(t, stored)

	memberships, err = datastore.ListMemberships(ctx)
	require.NoError(t, err)
	require.Empty(t, memberships)
}

func TestMembershipForeignKeysConstraints(t *testing.T) {
	t.Parallel()
	datastore, err := setupDatastore(t)
	require.NoError(t, err, "Failed to setup the datastore")

	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)

	// Create Federation Group
	fg1 := &entity.FederationGroup{
		Name:        "fg1-test",
		Description: "test-federation-group",
		Status:      entity.StatusActive,
	}
	fg1, err = datastore.CreateOrUpdateFederationGroup(ctx, fg1)
	require.NoError(t, err)

	m1 := &entity.Member{
		TrustDomain: td1,
		Status:      entity.StatusActive,
	}
	m1, err = datastore.CreateOrUpdateMember(ctx, m1)
	require.NoError(t, err)

	membership1 := &entity.Membership{
		MemberID:          m1.ID,
		FederationGroupID: fg1.ID,
		Status:            entity.StatusPending,
	}
	membership1, err = datastore.CreateOrUpdateMembership(ctx, membership1)
	require.NoError(t, err)

	// Cannot add a new membership for the same Member and Federation Group
	membership1.ID = uuid.NullUUID{}
	_, err = datastore.CreateOrUpdateMembership(ctx, membership1)
	require.Error(t, err)

	wrappedErr := errors.Unwrap(err)
	errCode := wrappedErr.(*pgconn.PgError).SQLState()
	assert.Equal(t, pgerrcode.UniqueViolation, errCode, "Unique constraint violation error was expected")

	// Cannot delete Federation Group that has a membership associated
	err = datastore.DeleteFederationGroup(ctx, fg1.ID.UUID)
	require.Error(t, err)

	wrappedErr = errors.Unwrap(err)
	errCode = wrappedErr.(*pgconn.PgError).SQLState()
	assert.Equal(t, pgerrcode.ForeignKeyViolation, errCode, "Foreign key violation error was expected")

	// Cannot delete Member that has a membership associated
	err = datastore.DeleteMember(ctx, m1.ID.UUID)
	require.Error(t, err)

	wrappedErr = errors.Unwrap(err)
	errCode = wrappedErr.(*pgconn.PgError).SQLState()
	assert.Equal(t, pgerrcode.ForeignKeyViolation, errCode, "Foreign key violation error was expected")
}

func TestCRUDBundle(t *testing.T) {
	t.Parallel()
	datastore, err := setupDatastore(t)
	require.NoError(t, err, "Failed to setup the datastore")

	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)

	// Create members to associate the bundles
	m1 := &entity.Member{
		TrustDomain: td1,
		Status:      entity.StatusActive,
	}
	m1, err = datastore.CreateOrUpdateMember(ctx, m1)
	require.NoError(t, err)
	require.NotNil(t, m1.ID)

	m2 := &entity.Member{
		TrustDomain: td2,
		Status:      entity.StatusActive,
	}
	m2, err = datastore.CreateOrUpdateMember(ctx, m2)
	require.NoError(t, err)
	require.NotNil(t, m2.ID)

	// Create first Bundle - member-1
	req1 := &entity.Bundle{
		RawBundle:    []byte{1, 2, 3},
		Digest:       []byte{10, 20, 30},
		SignedBundle: []byte{4, 2},
		TlogID: uuid.NullUUID{
			UUID:  uuid.New(),
			Valid: true,
		},
		SvidPem:  "svid-pem",
		MemberID: m1.ID,
	}

	b1, err := datastore.CreateOrUpdateBundle(ctx, req1)
	require.NoError(t, err)
	require.NotNil(t, b1)
	assert.Equal(t, req1.RawBundle, b1.RawBundle)
	assert.Equal(t, req1.Digest, b1.Digest)
	assert.Equal(t, req1.SignedBundle, b1.SignedBundle)
	assert.Equal(t, req1.TlogID, b1.TlogID)
	assert.Equal(t, req1.SvidPem, b1.SvidPem)
	assert.Equal(t, req1.MemberID, b1.MemberID)

	// Look up bundle stored in DB and compare
	stored, err := datastore.FindBundleByID(ctx, b1.ID.UUID)
	require.NoError(t, err)
	assert.Equal(t, b1, stored)

	// Create second Bundle -> member-2
	req2 := &entity.Bundle{
		RawBundle:    []byte{4, 5, 6},
		Digest:       []byte{40, 50, 60},
		SignedBundle: []byte{7, 9},
		TlogID: uuid.NullUUID{
			UUID:  uuid.New(),
			Valid: true,
		},
		SvidPem:  "svid-2-pem",
		MemberID: m2.ID,
	}

	b2, err := datastore.CreateOrUpdateBundle(ctx, req2)
	require.NoError(t, err)
	require.NotNil(t, b1)
	assert.Equal(t, req2.RawBundle, b2.RawBundle)
	assert.Equal(t, req2.Digest, b2.Digest)
	assert.Equal(t, req2.SignedBundle, b2.SignedBundle)
	assert.Equal(t, req2.TlogID, b2.TlogID)
	assert.Equal(t, req2.SvidPem, b2.SvidPem)
	assert.Equal(t, req2.MemberID, b2.MemberID)

	// Look up bundle stored in DB and compare
	stored, err = datastore.FindBundleByID(ctx, b2.ID.UUID)
	require.NoError(t, err)
	assert.Equal(t, b2, stored)

	// Find bundles by MemberID
	stored, err = datastore.FindBundleByMemberID(ctx, m1.ID.UUID)
	require.NoError(t, err)
	assert.Equal(t, b1, stored)

	stored, err = datastore.FindBundleByMemberID(ctx, m2.ID.UUID)
	require.NoError(t, err)
	assert.Equal(t, b2, stored)

	// Update Bundle
	b1.RawBundle = []byte{'a', 'b', 'c'}
	b1.Digest = []byte{'c', 'd', 'e'}
	b1.SignedBundle = []byte{'f', 'g', 'h'}
	b1.SvidPem = "other-svid-pem"
	b1.TlogID = uuid.NullUUID{
		UUID:  uuid.New(),
		Valid: true,
	}

	updated, err := datastore.CreateOrUpdateBundle(ctx, b1)
	require.NoError(t, err)
	require.NotNil(t, updated)
	assert.Equal(t, b1.RawBundle, updated.RawBundle)
	assert.Equal(t, b1.Digest, updated.Digest)
	assert.Equal(t, b1.SignedBundle, updated.SignedBundle)
	assert.Equal(t, b1.TlogID, updated.TlogID)
	assert.Equal(t, b1.SvidPem, updated.SvidPem)
	assert.Equal(t, b1.MemberID, updated.MemberID)

	// Look up bundle stored in DB and compare
	stored, err = datastore.FindBundleByID(ctx, b1.ID.UUID)
	require.NoError(t, err)
	assert.Equal(t, updated, stored)

	// List bundles
	bundles, err := datastore.ListBundles(ctx)
	require.NoError(t, err)
	assert.Equal(t, 2, len(bundles))
	require.Contains(t, bundles, updated)
	require.Contains(t, bundles, b2)

	// Delete first bundle
	err = datastore.DeleteBundle(ctx, b1.ID.UUID)
	require.NoError(t, err)

	// Look up deleted bundle
	stored, err = datastore.FindBundleByID(ctx, b1.ID.UUID)
	require.NoError(t, err)
	require.Nil(t, stored)

	// Delete second bundle
	err = datastore.DeleteBundle(ctx, b2.ID.UUID)
	require.NoError(t, err)

	// Look up deleted bundle
	stored, err = datastore.FindBundleByID(ctx, b2.ID.UUID)
	require.NoError(t, err)
	require.Nil(t, stored)
}

func TestBundleUniqueMemberConstraint(t *testing.T) {
	t.Parallel()
	datastore, err := setupDatastore(t)
	require.NoError(t, err, "Failed to setup the datastore")

	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)

	// Create member to associate the bundles
	m1 := &entity.Member{
		TrustDomain: td1,
		Status:      entity.StatusActive,
	}
	m1, err = datastore.CreateOrUpdateMember(ctx, m1)
	require.NoError(t, err)
	require.NotNil(t, m1.ID)

	// Create Bundle
	b1 := &entity.Bundle{
		RawBundle:    []byte{1, 2, 3},
		Digest:       []byte{10, 20, 30},
		SignedBundle: []byte{4, 2},
		TlogID: uuid.NullUUID{
			UUID:  uuid.New(),
			Valid: true,
		},
		SvidPem:  "svid-pem",
		MemberID: m1.ID,
	}
	b1, err = datastore.CreateOrUpdateBundle(ctx, b1)
	require.NoError(t, err)
	require.NotNil(t, b1)

	// Create second Bundle associated to same member
	b2 := &entity.Bundle{
		RawBundle:    []byte{123, 212, 230},
		Digest:       []byte{120, 210, 130},
		SignedBundle: []byte{4, 2},
		TlogID: uuid.NullUUID{
			UUID:  uuid.New(),
			Valid: true,
		},
		SvidPem:  "svid-pem",
		MemberID: m1.ID,
	}
	b2, err = datastore.CreateOrUpdateBundle(ctx, b2)
	require.Error(t, err)
	require.Nil(t, b2)

	wrappedErr := errors.Unwrap(err)
	errCode := wrappedErr.(*pgconn.PgError).SQLState()
	assert.Equal(t, pgerrcode.UniqueViolation, errCode, "Unique constraint violation error was expected")
}

func TestCRUDJoinToken(t *testing.T) {
	t.Parallel()
	datastore, err := setupDatastore(t)
	require.NoError(t, err, "Failed to setup the datastore")

	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)

	//// Create members to associate the join tokens
	m1 := &entity.Member{
		TrustDomain: td1,
		Status:      entity.StatusActive,
	}
	m1, err = datastore.CreateOrUpdateMember(ctx, m1)
	require.NoError(t, err)
	require.NotNil(t, m1.ID)

	m2 := &entity.Member{
		TrustDomain: td2,
		Status:      entity.StatusActive,
	}
	m2, err = datastore.CreateOrUpdateMember(ctx, m2)
	require.NoError(t, err)
	require.NotNil(t, m2.ID)

	loc, _ := time.LoadLocation("UTC")
	expiry := time.Now().In(loc).Add(1 * time.Hour)

	// Create first join_token -> member_1
	req1 := &entity.JoinToken{
		Token:    uuid.NewString(),
		Expiry:   expiry,
		MemberID: m1.ID,
	}

	token1, err := datastore.CreateJoinToken(ctx, req1)
	require.NoError(t, err)
	require.NotNil(t, token1)
	assert.Equal(t, req1.Token, token1.Token)
	requireEqualDate(t, req1.Expiry, token1.Expiry.In(loc))
	require.False(t, token1.Used)
	assert.Equal(t, req1.MemberID, token1.MemberID)

	// Look up token stored in DB and compare
	stored, err := datastore.FindJoinTokensByID(ctx, token1.ID.UUID)
	require.NoError(t, err)
	assert.Equal(t, token1, stored)

	// Create second join_token -> member_2
	req2 := &entity.JoinToken{
		Token:    uuid.NewString(),
		Expiry:   expiry,
		MemberID: m2.ID,
	}

	token2, err := datastore.CreateJoinToken(ctx, req2)
	require.NoError(t, err)
	require.NotNil(t, token1)
	assert.Equal(t, req2.Token, token2.Token)
	assert.Equal(t, req1.MemberID, token1.MemberID)
	require.False(t, token2.Used)

	requireEqualDate(t, req2.Expiry, token2.Expiry.In(loc))
	assert.Equal(t, req2.MemberID, token2.MemberID)

	// Look up token stored in DB and compare
	stored, err = datastore.FindJoinTokensByID(ctx, token2.ID.UUID)
	require.NoError(t, err)
	assert.Equal(t, token2, stored)

	// Create second join_token -> member_2
	req3 := &entity.JoinToken{
		Token:    uuid.NewString(),
		Expiry:   expiry,
		MemberID: m2.ID,
	}

	token3, err := datastore.CreateJoinToken(ctx, req3)
	require.NoError(t, err)
	require.NotNil(t, token3)

	// Find tokens by MemberID
	tokens, err := datastore.FindJoinTokensByMemberID(ctx, m1.ID.UUID)
	require.NoError(t, err)
	assert.Equal(t, 1, len(tokens))
	require.Contains(t, tokens, token1)

	tokens, err = datastore.FindJoinTokensByMemberID(ctx, m2.ID.UUID)
	require.NoError(t, err)
	assert.Equal(t, 2, len(tokens))
	require.Contains(t, tokens, token2)
	require.Contains(t, tokens, token3)

	// Look up join token by token string
	stored, err = datastore.FindJoinToken(ctx, token1.Token)
	require.NoError(t, err)
	assert.Equal(t, token1, stored)

	stored, err = datastore.FindJoinToken(ctx, token2.Token)
	require.NoError(t, err)
	assert.Equal(t, token2, stored)

	stored, err = datastore.FindJoinToken(ctx, token3.Token)
	require.NoError(t, err)
	assert.Equal(t, token3, stored)

	// List tokens
	tokens, err = datastore.ListJoinTokens(ctx)
	require.NoError(t, err)
	assert.Equal(t, 3, len(tokens))
	require.Contains(t, tokens, token1)
	require.Contains(t, tokens, token2)
	require.Contains(t, tokens, token3)

	// Update join token
	updated, err := datastore.UpdateJoinToken(ctx, token1.ID.UUID, true)
	require.NoError(t, err)
	assert.Equal(t, true, updated.Used)

	// Look up and compare
	stored, err = datastore.FindJoinTokensByID(ctx, token1.ID.UUID)
	require.NoError(t, err)
	assert.Equal(t, true, stored.Used)
	assert.Equal(t, updated.UpdatedAt, stored.UpdatedAt)

	// Delete join tokens
	err = datastore.DeleteJoinToken(ctx, token1.ID.UUID)
	require.NoError(t, err)
	stored, err = datastore.FindJoinTokensByID(ctx, token1.ID.UUID)
	require.NoError(t, err)
	require.Nil(t, stored)

	err = datastore.DeleteJoinToken(ctx, token2.ID.UUID)
	require.NoError(t, err)
	stored, err = datastore.FindJoinTokensByID(ctx, token2.ID.UUID)
	require.NoError(t, err)
	require.Nil(t, stored)

	err = datastore.DeleteJoinToken(ctx, token3.ID.UUID)
	require.NoError(t, err)
	stored, err = datastore.FindJoinTokensByID(ctx, token3.ID.UUID)
	require.NoError(t, err)
	require.Nil(t, stored)

	tokens, err = datastore.ListJoinTokens(ctx)
	require.NoError(t, err)
	assert.Equal(t, 0, len(tokens))
}

func TestCRUDHarvester(t *testing.T) {
	t.Parallel()
	datastore, err := setupDatastore(t)
	require.NoError(t, err, "Failed to setup the datastore")

	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)

	//// Create members to associate the harvester
	m1 := &entity.Member{
		TrustDomain: td1,
		Status:      entity.StatusActive,
	}
	m1, err = datastore.CreateOrUpdateMember(ctx, m1)
	require.NoError(t, err)
	require.NotNil(t, m1.ID)

	m2 := &entity.Member{
		TrustDomain: td2,
		Status:      entity.StatusActive,
	}
	m2, err = datastore.CreateOrUpdateMember(ctx, m2)
	require.NoError(t, err)
	require.NotNil(t, m2.ID)

	// Create first harvester -> member_1
	req1 := &entity.Harvester{
		MemberID: m1.ID,
		IsLeader: false,
	}

	harvester1, err := datastore.CreateOrUpdateHarvester(ctx, req1)
	require.NoError(t, err)
	require.NotNil(t, harvester1)
	assert.Equal(t, req1.MemberID, harvester1.MemberID)
	assert.Equal(t, req1.IsLeader, harvester1.IsLeader)
	require.True(t, req1.LeaderUntil.Equal(harvester1.LeaderUntil))

	// Look up token stored in DB and compare
	stored, err := datastore.FindHarvesterByID(ctx, harvester1.ID.UUID)
	require.NoError(t, err)
	assert.Equal(t, harvester1, stored)

	loc, _ := time.LoadLocation("UTC")
	inOneHour := time.Now().Add(1 * time.Hour).In(loc)

	// Create second harvester -> member_2
	req2 := &entity.Harvester{
		MemberID:    m2.ID,
		IsLeader:    true,
		LeaderUntil: inOneHour,
	}

	harvester2, err := datastore.CreateOrUpdateHarvester(ctx, req2)
	require.NoError(t, err)
	require.NotNil(t, harvester1)
	assert.Equal(t, req2.MemberID, harvester2.MemberID)
	assert.Equal(t, req2.IsLeader, harvester2.IsLeader)
	requireEqualDate(t, req2.LeaderUntil, harvester2.LeaderUntil.In(loc))

	// Look up token stored in DB and compare
	stored, err = datastore.FindHarvesterByID(ctx, harvester2.ID.UUID)
	require.NoError(t, err)
	assert.Equal(t, harvester2, stored)

	// Create third harvester -> member_2
	req3 := &entity.Harvester{
		MemberID: m2.ID,
		IsLeader: false,
	}

	harvester3, err := datastore.CreateOrUpdateHarvester(ctx, req3)
	require.NoError(t, err)
	require.NotNil(t, harvester2)
	assert.Equal(t, req3.MemberID, harvester3.MemberID)
	assert.Equal(t, req3.IsLeader, harvester3.IsLeader)
	require.True(t, req3.LeaderUntil.Equal(harvester3.LeaderUntil))

	// Look up token stored in DB and compare
	stored, err = datastore.FindHarvesterByID(ctx, harvester3.ID.UUID)
	require.NoError(t, err)
	assert.Equal(t, harvester3, stored)

	// Find harvesters by MemberID
	harvesters, err := datastore.FindHarvestersByMemberID(ctx, m1.ID.UUID)
	require.NoError(t, err)
	assert.Equal(t, 1, len(harvesters))
	assert.Equal(t, harvester1, harvesters[0])

	harvesters, err = datastore.FindHarvestersByMemberID(ctx, m2.ID.UUID)
	require.NoError(t, err)
	assert.Equal(t, 2, len(harvesters))
	require.Contains(t, harvesters, harvester2)
	require.Contains(t, harvesters, harvester3)

	// Update Harvester
	harvester1.IsLeader = true
	harvester1.LeaderUntil = inOneHour
	updated, err := datastore.CreateOrUpdateHarvester(ctx, harvester1)
	require.NoError(t, err)
	assert.Equal(t, harvester1.IsLeader, updated.IsLeader)
	require.True(t, harvester1.LeaderUntil.Equal(updated.LeaderUntil))
	harvester1 = updated

	// Look up in DB and compare
	stored, err = datastore.FindHarvesterByID(ctx, harvester1.ID.UUID)
	require.NoError(t, err)
	assert.Equal(t, updated, stored)

	// List harvesters
	harvesters, err = datastore.ListHarvesters(ctx)
	require.NoError(t, err)
	assert.Equal(t, 3, len(harvesters))
	require.Contains(t, harvesters, harvester1)
	require.Contains(t, harvesters, harvester2)
	require.Contains(t, harvesters, harvester3)

	// Delete harvesters
	err = datastore.DeleteHarvester(ctx, harvester1.ID.UUID)
	require.NoError(t, err)
	stored, err = datastore.FindHarvesterByID(ctx, harvester1.ID.UUID)
	require.NoError(t, err)
	require.Nil(t, stored)

	err = datastore.DeleteHarvester(ctx, harvester2.ID.UUID)
	require.NoError(t, err)
	stored, err = datastore.FindHarvesterByID(ctx, harvester2.ID.UUID)
	require.NoError(t, err)
	require.Nil(t, stored)

	err = datastore.DeleteHarvester(ctx, harvester3.ID.UUID)
	require.NoError(t, err)
	stored, err = datastore.FindHarvesterByID(ctx, harvester3.ID.UUID)
	require.NoError(t, err)
	require.Nil(t, stored)

	// List harvesters
	harvesters, err = datastore.ListHarvesters(ctx)
	require.NoError(t, err)
	assert.Equal(t, 0, len(harvesters))
}

func requireEqualDate(t *testing.T, time1 time.Time, time2 time.Time) {
	y1, m1, d1 := time1.Date()
	y2, m2, d2 := time2.Date()
	h1, mt1, s1 := time1.Clock()
	h2, mt2, s2 := time2.Clock()

	require.Equal(t, y1, y2, "Year doesn't match")
	require.Equal(t, m1, m2, "Month doesn't match")
	require.Equal(t, d1, d2, "Day doesn't match")
	require.Equal(t, h1, h2, "Hour doesn't match")
	require.Equal(t, mt1, mt2, "Minute doesn't match")
	require.Equal(t, s1, s2, "Seconds doesn't match")
}

func setupDatastore(t *testing.T) (*Datastore, error) {
	log := logrus.New()

	conn := startDB(t)
	datastore, err := NewDatastore(log, conn)
	require.NoError(t, err)

	t.Cleanup(func() {
		err = datastore.Close()
		require.NoError(t, err)
	})
	return datastore, nil
}

// starts a postgres DB in a docker container and returns the connection string
func startDB(tb testing.TB) string {
	pool, err := dockertest.NewPool("")
	require.NoError(tb, err)

	err = pool.Client.Ping()
	require.NoError(tb, err)

	// pulls an image, creates a container based on it and runs it
	resource, err := pool.RunWithOptions(&dockertest.RunOptions{
		Repository: "postgres",
		Tag:        postgresImage,
		Env: []string{
			"POSTGRES_PASSWORD=" + password,
			"POSTGRES_USER=" + user,
			"POSTGRES_DB=" + dbname,
		},
	}, func(config *docker.HostConfig) {
		// set AutoRemove to true so that stopped container goes away by itself
		config.AutoRemove = true
		config.RestartPolicy = docker.RestartPolicy{Name: "no"}
	})
	require.NoError(tb, err)

	hostAndPort := resource.GetHostPort("5432/tcp")
	databaseUrl := fmt.Sprintf("postgres://%s:%s@%s/%s?sslmode=disable", user, password, hostAndPort, dbname)

	tb.Logf("Connecting to a test database on url: %s", databaseUrl)

	// wait until db in container is ready using exponential backoff-retry
	pool.MaxWait = 60 * time.Second
	if err = pool.Retry(func() error {
		db, err := sql.Open("postgres", databaseUrl)
		if err != nil {
			return err
		}
		defer func() {
			cerr := db.Close()
			if err != nil {
				err = cerr
			}
		}()

		return db.Ping()
	}); err != nil {
		tb.Fatalf("Could not connect to docker: %s", err)
	}

	tb.Cleanup(func() {
		err = pool.Purge(resource)
		require.NoError(tb, err)
	})

	return databaseUrl
}
