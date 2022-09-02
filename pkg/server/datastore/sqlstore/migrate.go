package sqlstore

import (
	"fmt"

	"gorm.io/gorm"
)

const (
	latestSchemaVersion = 0
)

// This function checks the schema version and calls the migration code before initializing the DB if needed
func migrateDB(db *gorm.DB) (err error) {
	//Check if Table exists
	isNew := !db.Migrator().HasTable(&Migration{})
	if err := db.Error; err != nil {
		return err
	}
	//If the DB is new, create all tables
	if isNew {
		return initDB(db)
	}
	// Retrieve the Migration table
	migration := Migration{}
	if err := db.First(&migration).Error; err != nil {
		return err
	}
	// Compare the version with the latest - TODO
	if migration.Version != latestSchemaVersion {
		return fmt.Errorf("schema migration not implemented")
	}

	return nil
}

func initDB(db *gorm.DB) error {
	// Creates the Table for the Member Model
	if err := db.AutoMigrate(&Member{}); err != nil {
		return fmt.Errorf("migrate error: automigrate: %v", err)
	}
	// Creates the Table for the Membership Model
	if err := db.AutoMigrate(&Membership{}); err != nil {
		return fmt.Errorf("migrate error: automigrate: %v", err)
	}
	// Creates the Table for the Relationship Model
	if err := db.AutoMigrate(&Relationship{}); err != nil {
		return fmt.Errorf("migrate error: automigrate: %v", err)
	}
	// Creates the Table for the TrustBundle Model
	if err := db.AutoMigrate(&TrustBundle{}); err != nil {
		return fmt.Errorf("migrate error: automigrate: %v", err)
	} // Creates the Table for the DB version control Model
	if err := db.AutoMigrate(&Migration{}); err != nil {
		return fmt.Errorf("migrate error: automigrate: %v", err)
	}
	// Setting the current DB version
	if err := db.Assign(Migration{Version: latestSchemaVersion}).FirstOrCreate(&Migration{}).Error; err != nil {
		return err
	}
	return nil
}
