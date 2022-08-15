package sqlstore

import (
	"errors"
	"fmt"
	"time"

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
		return nil, fmt.Errorf("unsupported database_type: %s" + dbtype)
	}
	db, err = dialectvar.connect(connectionString)
	if err != nil {
		return nil, fmt.Errorf("error connecting to: %s", connectionString)
	}
	return db, nil
}

func CreateAllTablesinDB(db *gorm.DB) (err error) {
	err = CreateOrganizationTableInDB(db)
	if err != nil {
		return err
	}
	err = CreateBridgeTableInDB(db)
	if err != nil {
		return err
	}
	err = CreateMemberTableInDB(db)
	if err != nil {
		return err
	}
	err = CreateMembershipTableInDB(db)
	if err != nil {
		return err
	}
	err = CreateRelationshipTableInDB(db)
	if err != nil {
		return err
	}
	err = CreateTrustbundleTableInDB(db)
	if err != nil {
		return err
	}
	return nil
}

// Creates the Table for the Organization Model
// Returns Error if AutoMigrate fails
func CreateOrganizationTableInDB(db *gorm.DB) error {
	err := db.AutoMigrate(&Organization{})
	if err != nil {
		return fmt.Errorf("sqlstorage error: automigrate: %v", err)
	}
	return nil
}

// Creates the Table for the Bridge Model
// Returns Error if AutoMigrate fails
func CreateBridgeTableInDB(db *gorm.DB) error {
	err := db.AutoMigrate(&Bridge{})
	if err != nil {
		return fmt.Errorf("sqlstorage error: automigrate: %v", err)
	}
	return nil
}

// Creates the Table for the Member Model
// Returns Error if AutoMigrate fails
func CreateMemberTableInDB(db *gorm.DB) error {
	err := db.AutoMigrate(&Member{})
	if err != nil {
		return fmt.Errorf("sqlstorage error: automigrate: %v", err)
	}
	return nil
}

// Creates the Table for the Membership Model
// Returns Error if AutoMigrate fails
func CreateMembershipTableInDB(db *gorm.DB) error {
	err := db.AutoMigrate(&Membership{})
	if err != nil {
		return fmt.Errorf("sqlstore error: automigrate: %v", err)
	}
	return nil
}

// Creates the Table for the Relationship Model
// Returns Error if AutoMigrate fails
func CreateRelationshipTableInDB(db *gorm.DB) error {

	err := db.AutoMigrate(&Relationship{})
	if err != nil {
		return fmt.Errorf("sqlstore error: automigrate: %v", err)
	}
	return nil
}

// Creates the Table for the Trustbundle Model
func CreateTrustbundleTableInDB(db *gorm.DB) error {

	err := db.AutoMigrate(&TrustBundle{})
	if err != nil {
		return fmt.Errorf("sqlstore error: automigrate: %v", err)
	}
	return nil
}

// Insert a new Organization into the DB.
// Ignores and returns nil if entry already exists. Returns an error if creation fails
func CreateOrganization(db *gorm.DB, org *Organization) (*Organization, error) {

	err := db.Where(&org).FirstOrCreate(org).Error
	if err != nil {
		return nil, fmt.Errorf("sqlstore error: %v", err)
	}
	return org, nil
}

// Creates a new Bridge or ATB  from an Organization Name
// Ignores and returns nil if it already exists. Returns an error if creation fails
func CreateBridge(db *gorm.DB, br *Bridge, orgID uint) (*Bridge, error) {

	org, err := RetrieveOrganizationbyID(db, orgID)
	if err != nil {
		return nil, err
	}
	br.OrganizationID = org.ID // Fill in the OrgID for the bridge
	err = db.Where(&br).FirstOrCreate(br).Error
	if err != nil {
		return nil, fmt.Errorf("sqlstore error: %v", err)
	}
	return br, nil
}

// Creates a new Member
// Ignores and returns nil if entry already exists. Returns an error if creation fails
func CreateMember(db *gorm.DB, mem *Member) (*Member, error) {
	err := db.Where(mem).FirstOrCreate(mem).Error
	if err != nil {
		return nil, fmt.Errorf("sqlstore error: %v", err)
	}
	return mem, nil
}

// Creates a new Membership from a Member
// Ignores and returns nil if entry already exists. Returns an error if creation fails
func CreateMembership(db *gorm.DB, memb *Membership, memberID uint, bridgeID uint) (*Membership, error) {
	// Check if Member exists in DB
	mem, err := RetrieveMemberbyID(db, memberID)
	if err != nil {
		return nil, err
	}
	var br *Bridge
	// Check if Bridge exists in DB
	br, err = RetrieveBridgebyID(db, bridgeID)
	if err != nil {
		return nil, err
	}
	memb.MemberID = mem.ID // Fill in the BridgeID for the bridge
	memb.BridgeID = br.ID  // Fill in the BridgeID for the bridge
	err = db.Where(memb).FirstOrCreate(memb).Error
	if err != nil {
		return nil, fmt.Errorf("sqlstore error: %v", err)
	}
	return memb, nil
}

// Create a new Trustbundle from a Member
// Ignores and returns nil if entry already exists. Returns an error if creation fails
func CreateTrustBundle(db *gorm.DB, trust *TrustBundle, memberID uint) (*TrustBundle, error) {

	mem, err := RetrieveMemberbyID(db, memberID)
	if err != nil {
		return nil, err
	}
	trust.MemberID = mem.ID // Fill in the BridgeID for the bridge
	err = db.Where(trust).FirstOrCreate(trust).Error
	if err != nil {
		return nil, fmt.Errorf("sqlstore error: %v", err)
	}
	return trust, nil
}

// Adds a new Relation between Two SPIRE Servers in DB using description as reference for the IDs
// Ignores and returns nil if entry already exists. Returns an error if creation fails
func CreateRelationship(db *gorm.DB, newrelation *Relationship, sourceID uint, targetID uint) error {

	// Check if Member exists in DB
	sourcemember, err := RetrieveMemberbyID(db, sourceID)
	if err != nil {
		return err
	}
	var targetmember *Member
	// Check if Member exists in DB
	targetmember, err = RetrieveMemberbyID(db, targetID)
	if err != nil {
		return err
	}
	newrelation.MemberID = sourcemember.ID       // Fill in the Source MemberID (Foreign Key)
	newrelation.TargetMemberID = targetmember.ID // Fill in the Target MemberID
	err = db.Where(&newrelation).FirstOrCreate(&newrelation).Error
	if err != nil {
		return fmt.Errorf("sqlstore error: %v", err)
	}
	return nil
}

// retrieves an Organization from the Database by Name. returns an error if query fails
func RetrieveOrganizationbyID(db *gorm.DB, orgID uint) (*Organization, error) {
	var org *Organization = &Organization{}
	err := db.Where("id = ?", orgID).First(org).Error
	if errors.Is(err, gorm.ErrRecordNotFound) { // If does not, throw an error
		return nil, fmt.Errorf("sqlstore error: organization %v does not exist in db", orgID)
	}
	if err != nil {
		return nil, fmt.Errorf("sqlstore error: %v", err)
	}
	return org, nil
}

// RetrieveBridgebyDescription retrieves a Bridge from the Database by description. returns an error if query fails
func RetrieveBridgebyID(db *gorm.DB, brID uint) (*Bridge, error) {
	var br *Bridge = &Bridge{}
	err := db.Where("id = ?", brID).First(br).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, fmt.Errorf("sqlstore error: bridge ID %v does not exist in db", brID)
	}
	if err != nil {
		return nil, fmt.Errorf("sqlstore error: %v", err)
	}
	return br, nil
}

// Retrieves all Bridges from the Database using Organization ID as reference. returns an error if the query fails
func RetrieveAllBridgesbyOrgID(db *gorm.DB, orgID uint) (*[]Bridge, error) {
	var org *Organization = &Organization{}
	err := db.Preload("Bridges").Where("ID = ?", orgID).Find(org).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, fmt.Errorf("sqlstore error: organization ID %d does not exist in db", orgID)
	}
	if err != nil {
		return nil, fmt.Errorf("sqlstore error: %v", err)
	}
	return &org.Bridges, nil
}

// Retrieves all Members from the Database using bridge ID as reference. returns an error if the query fails
func RetrieveAllMembershipsbyBridgeID(db *gorm.DB, bridgeID uint) (*[]Membership, error) {
	var br *Bridge = &Bridge{}
	err := db.Preload("Memberships").Where("ID = ?", bridgeID).Find(br).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, fmt.Errorf("sqlstore error: bridge %d does not exist in db", bridgeID)
	}
	if err != nil {
		return nil, fmt.Errorf("sqlstore error: %v. bridge query failed with bridge ID %d", err, bridgeID)
	}
	return &br.Memberships, nil
}

// Retrieves all Members from the Database using bridge ID as reference. returns an error if the query fails
func RetrieveAllMembersbyBridgeID(db *gorm.DB, bridgeID uint) (mem *[]Member, err error) {
	var br *Bridge = &Bridge{}
	err = db.Preload("Memberships.member").Where("ID = ?", bridgeID).Find(br).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, fmt.Errorf("sqlstore error: bridge %d does not exist in db", bridgeID)
	}
	if err != nil {
		return nil, fmt.Errorf("sqlstore error: %v. bridge query failed with bridge ID %d", err, bridgeID)
	}
	for _, membership := range br.Memberships {
		*mem = append(*mem, membership.member)
	}
	return mem, nil
}

// Retrieves all Members from the Database using bridge ID as reference. returns an error if the query fails
func RetrieveAllBridgesbyMemberID(db *gorm.DB, memberID uint) (mem *[]Bridge, err error) {
	var member *Member = &Member{}
	err = db.Preload("Memberships.bridge").Where("ID = ?", memberID).Find(member).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, fmt.Errorf("sqlstore error: bridge %d does not exist in db", memberID)
	}
	if err != nil {
		return nil, fmt.Errorf("sqlstore error: %v. bridge query failed with bridge ID %d", err, memberID)
	}
	for _, membership := range member.Memberships {
		*mem = append(*mem, membership.bridge)
	}
	return mem, nil
}

// Retrieves all Memberships from the Database using memberID as reference. returns an error if the query fails
func RetrieveAllMembershipsbyMemberID(db *gorm.DB, memberID uint) (*[]Membership, error) {
	var member *Member = &Member{}
	err := db.Preload("Memberships").Where("ID = ?", memberID).Find(member).Error
	if err != nil {
		return nil, fmt.Errorf("sqlstore error: %v. Membership query failed with member ID %d", err, memberID)
	}
	return &member.Memberships, nil
}

/// Retrieves all Relationships from the Database using memberID as reference. returns an error if the query fails
func RetrieveAllRelationshipsbyMemberID(db *gorm.DB, memberID uint) (*[]Relationship, error) {
	var member *Member = &Member{}
	err := db.Preload("Relationships").Where("ID = ?", memberID).Find(member).Error
	if err != nil {
		return nil, fmt.Errorf("sqlstore error: %v. relationship query failed with member ID %d", err, memberID)
	}
	return &member.Relationships, nil
}

// Retrieves all Trusts from the Database using memberID as reference. returns an error if the query fails
func RetrieveAllTrustBundlesbyMemberID(db *gorm.DB, memberID uint) (*[]TrustBundle, error) {
	var member *Member = &Member{}
	err := db.Preload("TrustBundles").Where("ID = ?", memberID).Find(member).Error
	if err != nil {
		return nil, fmt.Errorf("sqlstore error: %v. trust query failed with member id %d", err, memberID)
	}
	return &member.TrustBundles, nil
}

// Retrieves a Member from the Database by description. returns an error if the query fails
func RetrieveMemberbyID(db *gorm.DB, memberID uint) (*Member, error) {
	var member *Member = &Member{}
	err := db.Where("id = ?", memberID).First(member).Error
	if errors.Is(err, gorm.ErrRecordNotFound) { // If does not, throw an error
		return nil, fmt.Errorf("sqlstore error: Member with ID=%v does not exist in DB", memberID)
	}
	if err != nil {
		return nil, fmt.Errorf("sqlstore error: %v", err)
	}
	return member, nil
}

// RetrieveMembershipbyToken retrieves a Membership from the Database bigger than an specific date. returns an error if something goes wrong.
func RetrieveMembershipbyCreationDate(db *gorm.DB, date time.Time) (*Membership, error) {
	var membership *Membership = &Membership{}
	err := db.Where("created_at >= ?", date).First(membership).Error
	if errors.Is(err, gorm.ErrRecordNotFound) { // If does not, throw an error
		return nil, fmt.Errorf("sqlstore error: member created_at=%v does not exist in DB", date)
	}
	if err != nil {
		return nil, fmt.Errorf("sqlstore error: %v", err)
	}
	return membership, nil
}

// RetrieveMembershipbyToken retrieves a Membership from the Database by Token. returns an error if something goes wrong.
func RetrieveMembershipbyToken(db *gorm.DB, token string) (*Membership, error) {
	var membership *Membership = &Membership{}
	err := db.Where("join_token = ?", token).First(membership).Error
	if errors.Is(err, gorm.ErrRecordNotFound) { // If does not, throw an error
		return nil, fmt.Errorf("sqlstore error: Member with Token=%v does not exist in DB", token)
	}
	if err != nil {
		return nil, fmt.Errorf("sqlstore error: %v", err)
	}
	return membership, nil
}

// retrieves a Relationship from the Database by Source and Target IDs. returns an error if something goes wrong.
func RetrieveRelationshipbySourceandTargetID(db *gorm.DB, source uint, target uint) (*Relationship, error) {
	var relationship *Relationship = &Relationship{}
	err := db.Where("MemberID = ? AND TargetMemberID = ?", source, target).First(relationship).Error
	if errors.Is(err, gorm.ErrRecordNotFound) { // If does not, throw an error
		return nil, fmt.Errorf("sqlstore error: Member with SourceMemberID=%v and/or TargetMemberID=%v does not exist in DB", source, target)
	}
	if err != nil {
		return nil, fmt.Errorf("sqlstore error: %v", err)
	}
	return relationship, nil
}

// retrieves a TrustBundle from the Database by Token. returns an error if something goes wrong.
func RetrieveTrustbundlebyMemberID(db *gorm.DB, memberID string) (*TrustBundle, error) {
	var trustbundle *TrustBundle = &TrustBundle{}
	err := db.Where("MemberID = ?", memberID).First(trustbundle).Error
	if errors.Is(err, gorm.ErrRecordNotFound) { // If does not, throw an error
		return nil, fmt.Errorf("sqlstore error: Member with Token=%v does not exist in DB", memberID)
	}
	if err != nil {
		return nil, fmt.Errorf("sqlstore error: %v", err)
	}
	return trustbundle, nil
}

// UpdateBridge Updates an existing Bridge with the new Bridge as argument. The ID will be used as reference.
func UpdateBridge(db *gorm.DB, br Bridge) error {
	if br.ID == 0 {
		return fmt.Errorf("sqlstore error: Bridge ID is invalid")
	}
	err := db.Model(&br).Updates(&br).Error
	if err != nil {
		return fmt.Errorf("sqlstore error: %v", err)
	}
	return nil
}

// Updates an existing Organization with the new Organization as argument. The ID will be used as reference.
func UpdateOrganization(db *gorm.DB, org Organization) error {
	if org.ID == 0 {
		return fmt.Errorf("sqlstore error: organization ID is invalid")
	}
	err := db.Model(&org).Updates(&org).Error
	if errors.Is(err, gorm.ErrRecordNotFound) { // If does not, throw an error
		return fmt.Errorf("sqlstore error: organization with ID %d does not exist", org.ID)
	}
	if err != nil {
		return fmt.Errorf("sqlstore error: %v", err)
	}
	return nil
}

// Updates an existing Member with the new Member as argument. The ID will be used as reference.
func UpdateMember(db *gorm.DB, member Member) error {
	if member.ID == 0 {
		return fmt.Errorf("sqlstore error: member id is invalid")
	}
	err := db.Model(&member).Updates(&member).Error
	if err != nil {
		return fmt.Errorf("sqlstore error: %v", err)
	}
	return nil
}

// Updates an existing Member with the new Member as argument. The ID will be used as reference.
func UpdateMembership(db *gorm.DB, membership Membership) error {
	if membership.ID == 0 {
		return fmt.Errorf("sqlstore error: membership id is invalid")
	}
	err := db.Model(&membership).Updates(&membership).Error
	if err != nil {
		return fmt.Errorf("sqlstore error: %v", err)
	}
	return nil
}

// Updates an existing Member with the new Member as argument. The ID will be used as reference.
func UpdateTrust(db *gorm.DB, trust TrustBundle) error {
	if trust.ID == 0 {
		return fmt.Errorf("sqlstore error: membership id is invalid")
	}
	err := db.Model(&trust).Updates(&trust).Error
	if err != nil {
		return fmt.Errorf("sqlstore error: %v", err)
	}
	return nil
}

// Delete Organization by name from the DB with cascading, returning error if something happens
func DeleteOrganizationbyID(db *gorm.DB, orgID uint) error {
	org, err := RetrieveOrganizationbyID(db, orgID)
	if err != nil {
		return err
	}

	if db.Name() == "sqlite" {
		// Workaround for https://github.com/mattn/go-sqlite3/pull/802 that
		// might prevent DELETE CASCADE on go-sqlite3 driver from working
		brs, err := RetrieveAllBridgesbyOrgID(db, orgID)
		if err != nil {
			return err
		}
		for _, br := range *brs {
			err = DeleteBridgebyID(db, br.ID)
			if err != nil {
				return err
			}
		}
	}
	err = db.Model(&org).Delete(&org).Error
	if err != nil {
		return fmt.Errorf("sqlstore error: %v", err)
	}
	return nil
}

// Delete Organization by Description from the DB with cascading
func DeleteBridgebyID(db *gorm.DB, bridgeID uint) error {
	br, err := RetrieveBridgebyID(db, bridgeID)
	if err != nil {
		return err
	}
	if db.Name() == "sqlite" {
		// Workaround for https://github.com/mattn/go-sqlite3/pull/802 that
		// might prevent DELETE CASCADE on go-sqlite3 driver from working
		memberships, err := RetrieveAllMembershipsbyBridgeID(db, bridgeID)
		if err != nil {
			return err
		}
		for _, membership := range *memberships {
			err = DeleteMembershipbyToken(db, membership.JoinToken)
			if err != nil {
				return err
			}
		}
	}
	// Deletes the Bridge. If its MySQL or Postgres it will cascade automatically by DB model constraint
	err = db.Model(&br).Delete(&br).Error
	if err != nil {
		return fmt.Errorf("sqlstore error: %v", err)
	}
	return nil
}

// Delete Organization by name from the DB with cascading
func DeleteMemberbyID(db *gorm.DB, memberID uint) error {
	member, err := RetrieveMemberbyID(db, memberID)
	if err != nil {
		return fmt.Errorf("sqlstore error: %v", err)
	}
	if db.Name() == "sqlite" {
		// Workaround for https://github.com/mattn/go-sqlite3/pull/802 that
		// might prevent DELETE CASCADE on go-sqlite3 driver from working
		err = DeleteAllMembershipsbyMemberID(db, memberID)
		if err != nil {
			return err
		}
		err = DeleteAllRelationshipsbyMemberID(db, memberID)
		if err != nil {
			return err
		}
		err = DeleteAllTrustbundlesbyMemberID(db, memberID)
		if err != nil {
			return err
		}

	}
	// Deletes the Member. If its MySQL or Postgres it will cascade automatically by DB model constraint
	err = db.Model(&member).Delete(&member).Error
	if err != nil {
		return fmt.Errorf("sqlstore error: %v", err)
	}
	return nil
}

// Deletes all Memberships using memberid as FK
func DeleteAllMembershipsbyMemberID(db *gorm.DB, memberid uint) error {
	memberships, err := RetrieveAllMembershipsbyMemberID(db, memberid)
	if err != nil {
		return err
	}
	for _, membership := range *memberships {
		err = db.Model(&membership).Delete(&membership).Error
		if err != nil {
			return fmt.Errorf("sqlstore error: %v. Error deleting Relationships from member with id %d", err, memberid)
		}
	}
	return nil
}

// Deletes all Memberships using memberid as FK
func DeleteAllMembershipsbyBridgeID(db *gorm.DB, bridgeid uint) error {
	memberships, err := RetrieveAllMembershipsbyBridgeID(db, bridgeid)
	if err != nil {
		return err
	}
	for _, membership := range *memberships {
		err = db.Model(&membership).Delete(&membership).Error
		if err != nil {
			return fmt.Errorf("sqlstore error: %v. Error deleting Relationships from bridge with id %d", err, bridgeid)
		}
	}
	return nil
}

// Deletes all Relationships using memberid as FK
func DeleteAllRelationshipsbyMemberID(db *gorm.DB, memberid uint) error {
	relations, err := RetrieveAllRelationshipsbyMemberID(db, memberid)
	if err != nil {
		return err
	}
	for _, relation := range *relations {
		err = db.Model(&relation).Delete(&relation).Error
		if err != nil {
			return fmt.Errorf("sqlstore error: %v. Error deleting Relationships from member with id %d", err, memberid)
		}
	}
	return nil
}

// Deletes all Trusts using memberid as FK
func DeleteAllTrustbundlesbyMemberID(db *gorm.DB, memberid uint) error {
	trusts, err := RetrieveAllTrustBundlesbyMemberID(db, memberid)
	if err != nil {
		return err
	}
	for _, trust := range *trusts {
		err = db.Model(&trust).Delete(&trust).Error
		if err != nil {
			return fmt.Errorf("sqlstore error: %v. Not able to fully delete trustbundle %s", err, trust.TrustBundle)
		}
	}
	return nil
}

// Delete membership by Token from the DB
func DeleteMembershipbyToken(db *gorm.DB, name string) error {
	membership, err := RetrieveMembershipbyToken(db, name)
	if err != nil {
		return err
	}
	err = db.Model(&membership).Delete(&membership).Error
	if err != nil {
		return fmt.Errorf("sqlstore error: %v. Not able to fully delete membership  with token %s", err, membership.JoinToken)
	}
	return nil
}

// Delete Relationship by Source and Target IDs from the DB
func DeleteRelationshipbySourceTargetID(db *gorm.DB, source uint, target uint) error {
	relationship, err := RetrieveRelationshipbySourceandTargetID(db, source, target)
	if err != nil {
		return err
	}
	err = db.Model(&relationship).Delete(&relationship).Error
	if err != nil {
		return fmt.Errorf("sqlstore error: %v", err)
	}
	return nil
}

// Delete Trusts by MemberID from the DB
func DeleteTrustBundlebyMemberID(db *gorm.DB, memberID string) error {
	trust, err := RetrieveTrustbundlebyMemberID(db, memberID)
	if err != nil {
		return err
	}
	err = db.Model(&trust).Delete(&trust).Error
	if err != nil {
		return fmt.Errorf("sqlstore error: %v", err)
	}
	return nil
}
