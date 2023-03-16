package datastore

import (
	"database/sql"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/golang-migrate/migrate/v4/source"
)

// currentDBVersion defines the current migration version supported by the app.
// This is used to ensure that the app is compatible with the database schema.
// The current DBVersion, scheme, and source instance must be managed by each datastore.

func ValidateAndMigrateSchema(db *sql.DB, currentDBVersion uint, scheme string, sourceInstance source.Driver) error {

	driverInstance, err := postgres.WithInstance(db, new(postgres.Config))
	if err != nil {
		return err
	}

	m, err := migrate.NewWithInstance("iofs", sourceInstance, scheme, driverInstance)
	if err != nil {
		return err
	}

	err = m.Migrate(currentDBVersion)
	if err != nil && err != migrate.ErrNoChange {
		return err
	}

	return sourceInstance.Close()
}
