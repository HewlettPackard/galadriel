package sqlstore

import (
	"errors"
	"fmt"

	"gorm.io/gorm"
)

const (
	// MySQL database type
	MySQL = "mysql"
	// PostgreSQL database type
	PostgreSQL = "postgres"
	// SQLite database type
	SQLite = "sqlite3"
)

func OpenDB(connectionString, dbtype string) (*gorm.DB, error) {
	var dialectvar dialect

	switch dbtype {
	case SQLite:
		dialectvar = sqliteDB{}
	case PostgreSQL:
		dialectvar = postgresDB{}
	case MySQL:
		dialectvar = mysqlDB{}
	default:
		return nil, fmt.Errorf("unsupported database_type: %s", dbtype)
	}
	db, err := dialectvar.connect(connectionString)
	if err != nil {
		return nil, errors.New("Error connecting to: %s" + connectionString)
	}
	return db, nil
}

func CreateOrganizationTableInDB(db *gorm.DB) error {
	// Creates the Model for the Bridge
	result := db.AutoMigrate(&Organization{})
	if result != nil {
		return fmt.Errorf("sqlstore error: %v", result)
	}
	return nil
}

func InsertOrg(db *gorm.DB, org Organization) error {
	// Insert new ORG into the DB
	if err := db.Where(&org).FirstOrCreate(&org).Error; err != nil {
		return fmt.Errorf("sqlstorage error: %v", err)
	}
	return nil
}
// RetrieveOrg retrieves an Organization from the Database. returns an error if something goes wrong.
func RetrieveOrg(db *gorm.DB, org *Organization) error {
	result := db.Where("Name = ?", (*org).Name).First(org)
	if errors.Is(result.Error, gorm.ErrRecordNotFound) { // If does not, throw an error
		fmt.Println("storage err: ", gorm.ErrRecordNotFound)
		return errors.New("Organization " + (*org).Name + " does not exist in DB")
	}
	return nil
}

func UpdateOrg(db *gorm.DB, org Organization) error {
	// Insert new ORG into the DB
	if org.ID == 0 {
		return errors.New("OrgID is invalid")
	}
	err := db.Model(&org).Updates(&org).Error
	if err != nil {
		fmt.Println("storage err: ", err)
		return err
	}
	return nil
}

func CreateBridgeTableInDB(db *gorm.DB) error {
	// Creates the Model for the Bridge
	err := db.AutoMigrate(&Bridge{})
	if err != nil {
		fmt.Println("storage err: ", err)
		return err
	}
	return nil
}

func RetrieveBridge(db *gorm.DB, br *Bridge) error {

	err := db.Where(br).First(br).Error
	if err != nil {
		fmt.Println("storage err: ", err)
		return err
	}
	return nil
}
func InsertBridge(db *gorm.DB, br Bridge, org Organization) error {

	result := db.Where("Name = ?", org.Name).First(&org) //Search if the org exists
	if errors.Is(result.Error, gorm.ErrRecordNotFound) { // If does not, throw an error
		return errors.New("Organization " + org.Name + " does not exist in DB")
	}
	br.OrganizationID = org.ID // Fill in the OrgID for the bridge
	err := db.Where(&br).FirstOrCreate(&br).Error
	if err != nil {
		fmt.Println("storage err: ", err)
		return err
	}
	return nil
}

func UpdateBridge(db *gorm.DB, br Bridge) error {
	if br.ID == 0 {
		return errors.New("BridgeID is invalid")
	}
	err := db.Model(&br).Updates(&br).Error
	if err != nil {
		fmt.Println("storage err: ", err)
		return err
	}
	return nil
}

func CreateMemberTableInDB(db *gorm.DB) error {
	// Creates the Model for the Bridge
	result := db.AutoMigrate(&Member{})
	if result != nil {
		fmt.Println("storage err: ", result)
		return result
	}
	return nil
}
