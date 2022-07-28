package datastore

import (
	"errors"
	"fmt"

	"gorm.io/gorm"
)

func CreateOrganizationTableInDB(db *gorm.DB) error {
	// Creates the Model for the Bridge
	result := db.AutoMigrate(&Organization{})
	if result != nil {
		fmt.Println("storage err: ", result)
		return result
	}
	return nil
}

func InsertOrg(db *gorm.DB, org Organization) error {
	// Insert new ORG into the DB
	err := db.Where(&org).FirstOrCreate(&org).Error
	if err != nil {
		fmt.Println("storage err: ", err)
		return err
	}
	return nil
}

func CreateBridgeTableInDB(db *gorm.DB) error {
	// Creates the Model for the Bridge
	result := db.AutoMigrate(&Bridge{})
	if result != nil {
		fmt.Println("storage err: ", result)
		return result
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
