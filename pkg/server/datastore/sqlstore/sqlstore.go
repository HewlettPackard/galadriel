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

func OpenDB(connectionString, dbtype string) (db *gorm.DB, err error) {
	var dialectvar dialect

	switch dbtype {
	case SQLite:
		dialectvar = sqliteDB{}
	case PostgreSQL:
		dialectvar = postgresDB{}
	case MySQL:
		dialectvar = mysqlDB{}
	default:
		return nil, errors.New("unsupported database_type: %s" + dbtype)
	}
	db, err = dialectvar.connect(connectionString)
	if err != nil {
		return nil, fmt.Errorf("error connecting to: %s", connectionString)
	}
	return db, nil
}

func CreateOrganizationTableInDB(db *gorm.DB) error {
	// Creates the Table for the Organization
	err := db.AutoMigrate(&Organization{})
	if err != nil {
		return fmt.Errorf("sqlstorage error: %v", err)
	}
	return nil
}

func CreateBridgeTableInDB(db *gorm.DB) error {
	// Creates the Table for the Bridge
	err := db.AutoMigrate(&Bridge{})
	if err != nil {
		return fmt.Errorf("sqlstorage error: %v", err)
	}
	return nil
}

func CreateMemberTableInDB(db *gorm.DB) error {
	// Creates the Table for the Member
	err := db.AutoMigrate(&Member{})
	if err != nil {
		return fmt.Errorf("sqlstorage error: %v", err)
	}
	return nil
}

func CreateMembershipTableInDB(db *gorm.DB) error {
	// Creates the Table for the Membership
	err := db.AutoMigrate(&Membership{})
	if err != nil {
		return fmt.Errorf("sqlstore error: %v", err)
	}
	return nil
}

func CreateRelationshipTableInDB(db *gorm.DB) error {
	// Creates the Table for the Relationship
	err := db.AutoMigrate(&Relationship{})
	if err != nil {
		return fmt.Errorf("sqlstore error: %v", err)
	}
	return nil
}

func CreateTrustbundleTableInDB(db *gorm.DB) error {
	// Creates the Table for the Trustbundle
	err := db.AutoMigrate(&TrustBundle{})
	if err != nil {
		return fmt.Errorf("sqlstore error: %v", err)
	}
	return nil
}

func CreateOrg(db *gorm.DB, org Organization) error {
	// Insert new ORG into the DB
	err := db.Where(&org).FirstOrCreate(&org).Error
	if err != nil {
		return fmt.Errorf("sqlstore error: %v", err)
	}
	return nil
}

func CreateBridge(db *gorm.DB, br Bridge, org Organization) error {
	// Creates a new Bridge or ATB
	result := db.Where("Name = ?", org.Name).First(&org) //Search if the org exists
	if errors.Is(result.Error, gorm.ErrRecordNotFound) { // If does not, throw an error
		return errors.New("Organization " + org.Name + " does not exist in DB")
	}
	br.OrganizationID = org.ID // Fill in the OrgID for the bridge
	err := db.Where(&br).FirstOrCreate(&br).Error
	if err != nil {
		return fmt.Errorf("sqlstore error: %v", err)
	}
	return nil
}

func CreateMember(db *gorm.DB, mem Member, br Bridge) error {
	// Creates a new Member
	result := db.Where("Description = ?", br.Description).First(&br) //Search if the bridge exists
	if errors.Is(result.Error, gorm.ErrRecordNotFound) {             // If does not, throw an error
		return errors.New("Bridge Description=" + br.Description + " does not exist in DB")
	}
	mem.BridgeID = br.ID // Fill in the BridgeID for the bridge
	err := db.Where(&mem).FirstOrCreate(&mem).Error
	if err != nil {
		return fmt.Errorf("sqlstore error: %v", err)
	}
	return nil
}

func CreateMembership(db *gorm.DB, memb Membership, mem Member) error {
	// Creates a new Membership
	result := db.Where("SpiffeID = ?", mem.SpiffeID).First(&mem) //Search if the bridge exists
	if errors.Is(result.Error, gorm.ErrRecordNotFound) {         // If does not, throw an error
		return errors.New("Member SpiffeID=" + mem.SpiffeID + " does not exist in DB")
	}
	memb.MemberID = mem.ID // Fill in the BridgeID for the bridge
	err := db.Where(&memb).FirstOrCreate(&memb).Error
	if err != nil {
		return fmt.Errorf("sqlstore error: %v", err)
	}
	return nil
}

func CreateTrustBundle(db *gorm.DB, trust TrustBundle, mem Member) error {
	// Create e new Trustbundle
	result := db.Where("SpiffeID = ?", mem.SpiffeID).First(&mem) //Search if the bridge exists
	if errors.Is(result.Error, gorm.ErrRecordNotFound) {         // If does not, throw an error
		return errors.New("Member SpiffeID=" + mem.SpiffeID + " does not exist in DB")
	}
	trust.MemberID = mem.ID // Fill in the BridgeID for the bridge
	err := db.Where(&trust).FirstOrCreate(&trust).Error
	if err != nil {
		return fmt.Errorf("sqlstore error: %v", err)
	}
	return nil
}

func CreateRelationship(db *gorm.DB, newrelation Relationship, sourcemember Member, targetmember Member) error {
	// Adds a new Relation between Two SPIRE Servers in DB
	firstmember := db.Where("SpiffeID = ?", sourcemember.SpiffeID).First(&sourcemember) //Search if the Member exists
	if errors.Is(firstmember.Error, gorm.ErrRecordNotFound) {                           // If does not, throw an error
		return errors.New("Member SpiffeID=" + sourcemember.SpiffeID + " does not exist in DB")
	}
	secondmember := db.Where("SpiffeID = ?", targetmember.SpiffeID).First(&targetmember) //Search if the Member exists
	if errors.Is(secondmember.Error, gorm.ErrRecordNotFound) {                           // If does not, throw an error
		return errors.New("Member SpiffeID=" + targetmember.SpiffeID + " does not exist in DB")
	}
	newrelation.MemberID = sourcemember.ID       // Fill in the Source MemberID (Foreign Key)
	newrelation.TargetMemberID = targetmember.ID // Fill in the Target MemberID
	err := db.Where(&newrelation).FirstOrCreate(&newrelation).Error
	if err != nil {
		return fmt.Errorf("sqlstore error: %v", err)
	}
	return nil
}

func RetrieveOrgbyName(db *gorm.DB, name string) (*Organization, error) {
	// Fetch Organization by Name
	var org *Organization = &Organization{}
	result := db.Where("Name = ?", name).First(org)
	if errors.Is(result.Error, gorm.ErrRecordNotFound) { // If does not, throw an error
		return nil, fmt.Errorf("Organization %v does not exist in DB", name)
	}
	return org, nil
}

func RetrieveBridgebyDescription(db *gorm.DB, description string) (*Bridge, error) {
	// Fetch Bridge by Description
	var br *Bridge = &Bridge{}
	result := db.Where("Description = ?", description).First(br)
	if errors.Is(result.Error, gorm.ErrRecordNotFound) { // If does not, throw an error
		return nil, fmt.Errorf("Bridge %v does not exist in DB", description)
	}
	return br, nil
}

func RetrieveMemberbyDescription(db *gorm.DB, description string) (*Member, error) {
	// Fetch Member by Description
	var member *Member = &Member{}
	result := db.Where("Description = ?", description).First(member)
	if errors.Is(result.Error, gorm.ErrRecordNotFound) { // If does not, throw an error
		return nil, fmt.Errorf("Member with Description=%v does not exist in DB", description)
	}
	return member, nil
}

func UpdateBridge(db *gorm.DB, br Bridge) error {
	// Updates an existing Bridge
	if br.ID == 0 {
		return errors.New("BridgeID is invalid")
	}
	err := db.Model(&br).Updates(&br).Error
	if err != nil {
		return fmt.Errorf("sqlstore error: %v", err)
	}
	return nil
}

func UpdateOrg(db *gorm.DB, org Organization) error {
	// Update Org by name from DB
	if org.ID == 0 {
		return errors.New("OrgID is invalid")
	}
	err := db.Model(&org).Updates(&org).Error
	if err != nil {
		return fmt.Errorf("sqlstore error: %v", err)
	}
	return nil
}

func DeleteOrgbyName(db *gorm.DB, name string) error {
	// Delete Org by name from DB without cascade
	org, err := RetrieveOrgbyName(db, name)
	if err != nil {
		return fmt.Errorf("sqlstore error: %v", err)
	}
	db.Model(&org).Delete(&org)
	return nil
}
