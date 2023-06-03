package options

import "github.com/HewlettPackard/galadriel/pkg/common/entity"

type OrderDirection string

const (
	NotSet OrderDirection = ""
	Asc    OrderDirection = "ASC"
	Desc   OrderDirection = "DESC"
)

type ListRelationshipsCriteria struct {
	PageNumber            uint
	PageSize              uint
	FilterByConsentStatus *entity.ConsentStatus
	OrderByCreatedAt      OrderDirection
}
