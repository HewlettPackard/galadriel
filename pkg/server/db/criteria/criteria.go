// Package criteria provides criteria structures for filtering and ordering database queries.
package criteria

import (
	"github.com/HewlettPackard/galadriel/pkg/common/entity"
	"github.com/HewlettPackard/galadriel/pkg/server/db/dbtypes"
	"github.com/Masterminds/squirrel"
	"github.com/google/uuid"
)

// OrderDirection represents the direction for ordering results.
type OrderDirection string

const (
	NoOrder         OrderDirection = ""
	OrderAscending  OrderDirection = "asc"
	OrderDescending OrderDirection = "desc"
)

// Filter is an interface that defines a database filter.
type Filter interface {
	GetCondition(dbtypes.Engine) squirrel.Sqlizer
}

// QueryCriteria is an interface that defines criteria for filtering and ordering database queries.
type QueryCriteria interface {
	GetPageSize() uint
	GetPageNumber() uint
	GetOrderDirection() OrderDirection
	GetFilters() []Filter
}

// ConsentStatusFilter represents a filter based on consent status.
type ConsentStatusFilter struct {
	ConsentStatus *entity.ConsentStatus
}

// GetCondition returns the SQL condition based on the consent status filter.
func (f *ConsentStatusFilter) GetCondition(dbType dbtypes.Engine) squirrel.Sqlizer {
	if dbType == dbtypes.PostgreSQL {
		return squirrel.Or{
			squirrel.Expr("trust_domain_a_consent = $1 OR trust_domain_b_consent = $2", f.ConsentStatus, f.ConsentStatus),
		}
	}
	return squirrel.Or{
		squirrel.Eq{"trust_domain_a_consent": f.ConsentStatus},
		squirrel.Eq{"trust_domain_b_consent": f.ConsentStatus},
	}
}

// TrustDomainIDFilter represents a filter based on trust domain ID.
type TrustDomainIDFilter struct {
	TrustDomainID uuid.NullUUID
}

// GetCondition returns the SQL condition based on the trust domain ID filter.
func (f *TrustDomainIDFilter) GetCondition(dbType dbtypes.Engine) squirrel.Sqlizer {
	if dbType == dbtypes.PostgreSQL {
		return squirrel.Or{
			squirrel.Expr("trust_domain_a_id = $1 OR trust_domain_b_id = $2", f.TrustDomainID, f.TrustDomainID),
		}
	}
	return squirrel.Or{
		squirrel.Eq{"trust_domain_a_id": f.TrustDomainID},
		squirrel.Eq{"trust_domain_b_id": f.TrustDomainID},
	}
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

// GetFilters returns the filters for relationships.
func (c *ListRelationshipsCriteria) GetFilters() []Filter {
	var filters []Filter

	if c.FilterByConsentStatus != nil {
		filters = append(filters, &ConsentStatusFilter{ConsentStatus: c.FilterByConsentStatus})
	}

	if c.FilterByTrustDomainID.Valid {
		filters = append(filters, &TrustDomainIDFilter{TrustDomainID: c.FilterByTrustDomainID})
	}

	return filters
}

// ListTrustDomainsCriteria defines the criteria for filtering and ordering trust domains.
type ListTrustDomainsCriteria struct {
	PageNumber       uint           // Page number for pagination (0 for no pagination)
	PageSize         uint           // Number of items per page (0 for no pagination)
	OrderByCreatedAt OrderDirection // Order trust domains by created at (ascending, descending, or no order)
}

func (c *ListTrustDomainsCriteria) GetPageNumber() uint {
	return c.PageNumber
}

func (c *ListTrustDomainsCriteria) GetPageSize() uint {
	return c.PageSize
}

func (c *ListTrustDomainsCriteria) GetOrderDirection() OrderDirection {
	return c.OrderByCreatedAt
}

func (c *ListTrustDomainsCriteria) GetFilters() []Filter {
	return []Filter{}
}
