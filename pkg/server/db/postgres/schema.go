package postgres

import (
	"database/sql"
	"embed"

	"github.com/HewlettPackard/galadriel/pkg/server/db"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	"github.com/sirupsen/logrus"
)

//go:embed migrations/*.sql
var fs embed.FS

// supportedSchemaVersion defines the current Postgres schema migration version supported by the app.
// This is used to ensure that the app is compatible with the database schema.
// When a new migration is created, this version should be updated in order to force
// the migrations to run when starting up the app.
const supportedSchemaVersion = 2

const migrationsFolder = "migrations"

func applyMigrations(sqlDB *sql.DB, log logrus.FieldLogger) error {
	sourceInstance, err := iofs.New(fs, migrationsFolder)
	if err != nil {
		return err
	}
	defer sourceInstance.Close()

	driverInstance, err := postgres.WithInstance(sqlDB, new(postgres.Config))
	if err != nil {
		return err
	}

	migrated, err := db.ApplyDatabaseMigrations(driverInstance, sourceInstance, supportedSchemaVersion, driverName)
	if err != nil {
		return err
	}

	if migrated {
		log.Infof("Postgresql database migrated to version %d", supportedSchemaVersion)
	}

	return nil
}
