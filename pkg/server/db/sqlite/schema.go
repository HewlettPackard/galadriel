package sqlite

import (
	"database/sql"
	"embed"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/sqlite"
	"github.com/golang-migrate/migrate/v4/source/iofs"
)

//go:embed migrations/*.sql
var fs embed.FS

// currentDBVersion defines the current migration version supported by the app.
// This is used to ensure that the app is compatible with the database schema.
// When a new migration is created, this version should be updated in order to force
// the migrations to run when starting up the app.
const currentDBVersion = 1

const scheme = "sqlite3"

func validateAndMigrateSchema(db *sql.DB) error {

	sourceInstance, err := iofs.New(fs, "migrations")
	if err != nil {
		return err
	}

	driverInstance, err := sqlite.WithInstance(db, new(sqlite.Config))
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
