package tests

import (
	"database/sql"
	"fmt"
	"testing"
	"time"

	"github.com/HewlettPackard/galadriel/pkg/server/db/postgres"
	"github.com/HewlettPackard/galadriel/pkg/server/db/sqlite"
	"github.com/ory/dockertest/v3"
	"github.com/ory/dockertest/v3/docker"
	"github.com/stretchr/testify/require"
)

const (
	postgresImage = "15-alpine"
	user          = "test_user"
	password      = "test_password"
	dbname        = "test_db"
)

func setupSQLiteDatastore(t *testing.T) *sqlite.Datastore {
	// Use an in-memory database
	dsn := ":memory:"

	datastore, err := sqlite.NewDatastore(dsn)
	require.NoError(t, err)

	t.Cleanup(func() {
		err = datastore.Close()
		require.NoError(t, err)
	})

	return datastore
}

func setupPostgresDatastore(t *testing.T) *postgres.Datastore {
	conn := startPostgresDB(t)
	datastore, err := postgres.NewDatastore(conn)
	require.NoError(t, err)

	t.Cleanup(func() {
		err = datastore.Close()
		require.NoError(t, err)
	})
	return datastore
}

// starts a postgres DB in a docker container and returns the connection string
func startPostgresDB(tb testing.TB) string {
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
