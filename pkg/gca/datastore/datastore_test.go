package datastore_test

import (
	"context"
	"database/sql"
	"fmt"
	"testing"
	"time"

	"github.com/HewlettPackard/galadriel/pkg/gca/api"
	"github.com/HewlettPackard/galadriel/pkg/gca/datastore"
	"github.com/google/uuid"
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
	spiffeTD1 = spiffeid.RequireTrustDomainFromString("foo.test")
	spiffeTD2 = spiffeid.RequireTrustDomainFromString("bar.test")
)

func TestCreateTrustDomain(t *testing.T) {
	t.Parallel()
	ds, ctx := setupTest(t)

	req1 := &api.TrustDomain{
		Name: spiffeTD1,
	}
	td1, err := ds.CreateTrustDomain(ctx, req1)
	require.NoError(t, err)
	assert.Equal(t, req1.Name, td1.Name)
	require.NotNil(t, td1.ID)
	require.NotNil(t, td1.CreatedAt)
	require.NotNil(t, td1.UpdatedAt)

	// Look up trust domain stored in DB and compare
	stored, err := ds.FindTrustDomainByID(ctx, td1.ID.UUID)
	require.NoError(t, err)
	assert.Equal(t, td1, stored)
}

func TestUpdateTrustDomain(t *testing.T) {
	t.Parallel()
	ds, ctx := setupTest(t)

	req1 := &api.TrustDomain{
		Name: spiffeTD1,
	}
	td1, err := ds.CreateTrustDomain(ctx, req1)
	require.NoError(t, err)

	td1.Description = "updated_description"
	td1.HarvesterSpiffeID = spiffeid.RequireFromString("spiffe://domain/test")
	td1.OnboardingBundle = []byte{1, 2, 3}

	// Update Trust Domain
	updated1, err := ds.UpdateTrustDomain(ctx, td1)
	require.NoError(t, err)
	require.NotNil(t, updated1)

	// Look up trust domain stored in DB and compare
	stored, err := ds.FindTrustDomainByID(ctx, td1.ID.UUID)
	require.NoError(t, err)
	assert.Equal(t, td1.ID, stored.ID)
	assert.Equal(t, td1.Description, stored.Description)
	assert.Equal(t, td1.HarvesterSpiffeID, stored.HarvesterSpiffeID)
	assert.Equal(t, td1.OnboardingBundle, stored.OnboardingBundle)
}

func TestTrustFindDomainbyName(t *testing.T) {
	t.Parallel()
	ds, ctx := setupTest(t)

	req1 := &api.TrustDomain{
		Name: spiffeTD1,
	}
	td1, err := ds.CreateTrustDomain(ctx, req1)
	require.NoError(t, err)

	req2 := &api.TrustDomain{
		Name: spiffeTD2,
	}
	td2, err := ds.CreateTrustDomain(ctx, req2)
	require.NoError(t, err)

	stored1, err := ds.FindTrustDomainByName(ctx, td1.Name.String())
	require.NoError(t, err)
	assert.Equal(t, td1, stored1)

	stored2, err := ds.FindTrustDomainByName(ctx, td2.Name.String())
	require.NoError(t, err)
	assert.Equal(t, td2, stored2)
}

func TestDeleteTrustDomain(t *testing.T) {
	t.Parallel()
	ds, ctx := setupTest(t)

	req1 := &api.TrustDomain{
		Name: spiffeTD1,
	}
	td1, err := ds.CreateTrustDomain(ctx, req1)
	require.NoError(t, err)

	req2 := &api.TrustDomain{
		Name: spiffeTD2,
	}
	td2, err := ds.CreateTrustDomain(ctx, req2)
	require.NoError(t, err)

	err1 := ds.DeleteTrustDomain(ctx, td1.ID.UUID)
	require.NoError(t, err1)

	stored, err := ds.FindTrustDomainByID(ctx, td1.ID.UUID)
	require.NoError(t, err)
	require.Nil(t, stored)

	err = ds.DeleteTrustDomain(ctx, td2.ID.UUID)
	require.NoError(t, err)

	stored, err = ds.FindTrustDomainByID(ctx, td2.ID.UUID)
	require.NoError(t, err)
	require.Nil(t, stored)
}

func TestListTrustDomains(t *testing.T) {
	t.Parallel()
	ds, ctx := setupTest(t)

	req1 := &api.TrustDomain{
		Name: spiffeTD1,
	}
	td1, err := ds.CreateTrustDomain(ctx, req1)
	require.NoError(t, err)

	req2 := &api.TrustDomain{
		Name: spiffeTD2,
	}
	td2, err := ds.CreateTrustDomain(ctx, req2)
	require.NoError(t, err)

	list, err := ds.ListTrustDomains(ctx)
	require.NoError(t, err)
	require.NotNil(t, list)
	assert.Equal(t, 2, len(list))
	assert.Contains(t, list, td1)
	assert.Contains(t, list, td2)
}

func TestTrustDomainUniqueTrustDomainConstraint(t *testing.T) {
	t.Parallel()
	ds, ctx := setupTest(t)

	td1 := &api.TrustDomain{
		Name: spiffeTD1,
	}
	_, err := ds.CreateTrustDomain(ctx, td1)
	require.NoError(t, err)
	require.NotNil(t, td1.ID)

	// second trustDomain with same trust domain
	td2 := &api.TrustDomain{
		Name: spiffeTD1,
	}
	_, err = ds.CreateTrustDomain(ctx, td2)
	require.Error(t, err)

	//wrappedErr := errors.Unwrap(err)
	//errCode := wrappedErr.(*pgconn.PgError).SQLState()
	//assert.Equal(t, pgerrcode.UniqueViolation, errCode, "Unique constraint violation was expected")
}

func TestCRUDJoinToken(t *testing.T) {
	t.Parallel()
	ds, ctx := setupTest(t)
	// // Create trustDomains to associate the join tokens
	req1 := &api.TrustDomain{
		Name: spiffeTD1,
	}
	td1, err := ds.CreateTrustDomain(ctx, req1)
	require.NoError(t, err)
	require.NotNil(t, td1.ID)

	req2 := &api.TrustDomain{
		Name: spiffeTD2,
	}
	td2, err := ds.CreateTrustDomain(ctx, req2)
	require.NoError(t, err)
	require.NotNil(t, td2.ID)

	loc, _ := time.LoadLocation("UTC")
	expiry := time.Now().In(loc).Add(1 * time.Hour)

	// Create first join_token -> trustDomain_1
	reqjt1 := &api.JoinToken{
		Token:         uuid.NewString(),
		ExpiresAt:     expiry,
		TrustDomainID: td1.ID.UUID,
	}

	token1, err := ds.CreateJoinToken(ctx, reqjt1)
	require.NoError(t, err)
	require.NotNil(t, token1)
	assert.Equal(t, reqjt1.Token, token1.Token)
	assertEqualDate(t, reqjt1.ExpiresAt, token1.ExpiresAt.In(loc))
	require.False(t, token1.Used)
	assert.Equal(t, reqjt1.TrustDomainID, token1.TrustDomainID)

	// Look up token stored in DB and compare
	stored, err := ds.FindJoinTokenByID(ctx, token1.ID.UUID)
	require.NoError(t, err)
	assert.Equal(t, token1, stored)

	// Create second join_token -> trustDomain_2
	reqjt2 := &api.JoinToken{
		Token:         uuid.NewString(),
		ExpiresAt:     expiry,
		TrustDomainID: td2.ID.UUID,
	}

	token2, err := ds.CreateJoinToken(ctx, reqjt2)
	require.NoError(t, err)
	require.NotNil(t, token1)
	assert.Equal(t, reqjt2.Token, token2.Token)
	assert.Equal(t, reqjt2.TrustDomainID, token2.TrustDomainID)
	require.False(t, token2.Used)

	assertEqualDate(t, reqjt2.ExpiresAt, token2.ExpiresAt.In(loc))
	assert.Equal(t, reqjt2.TrustDomainID, token2.TrustDomainID)

	// Look up token stored in DB and compare
	stored, err = ds.FindJoinTokenByID(ctx, token2.ID.UUID)
	require.NoError(t, err)
	assert.Equal(t, token2, stored)

	// Create second join_token -> trustDomain_2
	reqjt3 := &api.JoinToken{
		Token:         uuid.NewString(),
		ExpiresAt:     expiry,
		TrustDomainID: td2.ID.UUID,
	}

	token3, err := ds.CreateJoinToken(ctx, reqjt3)
	require.NoError(t, err)
	require.NotNil(t, token3)

	// Find tokens by TrustDomainID
	tokens, err := ds.FindJoinTokensByTrustDomainID(ctx, td1.ID.UUID)
	require.NoError(t, err)
	assert.Equal(t, 1, len(tokens))
	require.Contains(t, tokens, token1)

	tokens, err = ds.FindJoinTokensByTrustDomainID(ctx, td2.ID.UUID)
	require.NoError(t, err)
	assert.Equal(t, 2, len(tokens))
	require.Contains(t, tokens, token2)
	require.Contains(t, tokens, token3)

	// Look up join token by token string
	stored, err = ds.FindJoinToken(ctx, token1.Token)
	require.NoError(t, err)
	assert.Equal(t, token1, stored)

	stored, err = ds.FindJoinToken(ctx, token2.Token)
	require.NoError(t, err)
	assert.Equal(t, token2, stored)

	stored, err = ds.FindJoinToken(ctx, token3.Token)
	require.NoError(t, err)
	assert.Equal(t, token3, stored)

	// List tokens
	tokens, err = ds.ListJoinTokens(ctx)
	require.NoError(t, err)
	assert.Equal(t, 3, len(tokens))
	require.Contains(t, tokens, token1)
	require.Contains(t, tokens, token2)
	require.Contains(t, tokens, token3)

	// Update join token
	updated, err := ds.UpdateJoinToken(ctx, token1.ID.UUID, true)
	require.NoError(t, err)
	assert.Equal(t, true, updated.Used)

	// Look up and compare
	stored, err = ds.FindJoinTokenByID(ctx, token1.ID.UUID)
	require.NoError(t, err)
	assert.Equal(t, true, stored.Used)
	assert.Equal(t, updated.UpdatedAt, stored.UpdatedAt)

	// Delete join tokens
	err = ds.DeleteJoinToken(ctx, token1.ID.UUID)
	require.NoError(t, err)
	stored, err = ds.FindJoinTokenByID(ctx, token1.ID.UUID)
	require.NoError(t, err)
	require.Nil(t, stored)

	err = ds.DeleteJoinToken(ctx, token2.ID.UUID)
	require.NoError(t, err)
	stored, err = ds.FindJoinTokenByID(ctx, token2.ID.UUID)
	require.NoError(t, err)
	require.Nil(t, stored)

	err = ds.DeleteJoinToken(ctx, token3.ID.UUID)
	require.NoError(t, err)
	stored, err = ds.FindJoinTokenByID(ctx, token3.ID.UUID)
	require.NoError(t, err)
	require.Nil(t, stored)

	tokens, err = ds.ListJoinTokens(ctx)
	require.NoError(t, err)
	assert.Equal(t, 0, len(tokens))
}

func assertEqualDate(t *testing.T, time1 time.Time, time2 time.Time) {
	y1, td1, d1 := time1.Date()
	y2, td2, d2 := time2.Date()
	h1, mt1, s1 := time1.Clock()
	h2, mt2, s2 := time2.Clock()

	require.Equal(t, y1, y2, "Year doesn't match")
	require.Equal(t, td1, td2, "Month doesn't match")
	require.Equal(t, d1, d2, "Day doesn't match")
	require.Equal(t, h1, h2, "Hour doesn't match")
	require.Equal(t, mt1, mt2, "Minute doesn't match")
	require.Equal(t, s1, s2, "Seconds doesn't match")
}

func setupTest(t *testing.T) (*datastore.SQLDatastore, context.Context) {
	ds, err := setupDatastore(t)
	require.NoError(t, err, "Failed to setup the ds")

	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)

	return ds, ctx
}

func setupDatastore(t *testing.T) (*datastore.SQLDatastore, error) {
	log := logrus.New()

	conn := startDB(t)
	datastore, err := datastore.NewSQLDatastore(log, conn)
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
	databaseURL := fmt.Sprintf("postgres://%s:%s@%s/%s?sslmode=disable", user, password, hostAndPort, dbname)

	tb.Logf("Connecting to a test database on url: %s", databaseURL)

	// wait until db in container is ready using exponential backoff-retry
	pool.MaxWait = 60 * time.Second
	if err = pool.Retry(func() error {
		db, err := sql.Open("postgres", databaseURL)
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

	return databaseURL
}
