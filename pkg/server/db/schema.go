package db

import (
	"fmt"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database"
	"github.com/golang-migrate/migrate/v4/source"
)

const migrationSourceDriverName = "iofs"

// ApplyDatabaseMigrations applies necessary database migrations to ensure the schema matches the current application version.
// It uses the go-migrate library to apply migrations from the given sourceInstance to the database represented by driverInstance.
// The expected version of the database schema is provided by targetSchemaVersion.
// The dbDriverName is the name of the database driver in use (e.g., "postgres", "sqlite").
// It returns a boolean indicating whether any migrations were applied, and any error encountered.
func ApplyDatabaseMigrations(driverInstance database.Driver, sourceInstance source.Driver, targetSchemaVersion uint, dbDriverName string) (bool, error) {
	m, err := migrate.NewWithInstance(migrationSourceDriverName, sourceInstance, dbDriverName, driverInstance)
	if err != nil {
		return false, fmt.Errorf("failed creating migration instance: %w", err)
	}

	err = m.Migrate(targetSchemaVersion)
	if err != nil && err != migrate.ErrNoChange {
		return false, fmt.Errorf("failed applying migrations: %w", err)
	}

	return err != migrate.ErrNoChange, nil
}
