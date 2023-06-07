package criteria

import (
	"github.com/HewlettPackard/galadriel/pkg/common/entity"
	"github.com/google/uuid"
)

type OrderDirection string

const (
	NoOrder         OrderDirection = ""
	OrderAscending  OrderDirection = "asc"
	OrderDescending OrderDirection = "desc"
)

type Criteria interface {
	GetPageSize() uint
	GetPageNumber() uint
	GetOrderDirection() OrderDirection
}

// ListRelationshipsCriteria defines the criteria for filtering and ordering relationships.
// When both FilterByConsentStatus and FilterByTrustDomainID are set, the relationships returned will be the ones that have
// the consent status on the field corresponding to the trust domain ID. This means the relationships must match the consent
// status specified in either trust_domain_a_consent or trust_domain_b_consent field, depending on the specified trust domain ID.
// If only one of the filter criteria is set, the relationships will be filtered based on that criterion alone.
// If none of the filter criteria are set, all relationships will be returned without any filtering.
type ListRelationshipsCriteria struct {
	PageNumber            uint                  // Page number for pagination (0 for no pagination)
	PageSize              uint                  // Number of items per page (0 for no pagination)
	FilterByConsentStatus *entity.ConsentStatus // Filter relationships by consent status (optional)
	FilterByTrustDomainID uuid.NullUUID         // Filter relationships by trust domain ID (optional)
	OrderByCreatedAt      OrderDirection        // Order relationships by created at (ascending, descending, or no order)
}

func (c *ListRelationshipsCriteria) GetPageNumber() uint {
	return c.PageNumber
}

func (c *ListRelationshipsCriteria) GetPageSize() uint {
	return c.PageSize
}

func (c *ListRelationshipsCriteria) GetOrderDirection() OrderDirection {
	return c.OrderByCreatedAt
}

// ListTrustDomainCriteria defines the criteria for filtering and ordering trust domains.
type ListTrustDomainCriteria struct {
	PageNumber       uint           // Page number for pagination (0 for no pagination)
	PageSize         uint           // Number of items per page (0 for no pagination)
	OrderByCreatedAt OrderDirection // Order relationships by created at (ascending, descending, or no order)
}

func (c *ListTrustDomainCriteria) GetPageNumber() uint {
	return c.PageNumber
}

func (c *ListTrustDomainCriteria) GetPageSize() uint {
	return c.PageSize
}

func (c *ListTrustDomainCriteria) GetOrderDirection() OrderDirection {
	return c.OrderByCreatedAt
}
