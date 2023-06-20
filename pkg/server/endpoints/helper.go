package endpoints

import (
	"fmt"

	"github.com/HewlettPackard/galadriel/pkg/common/api"
	"github.com/HewlettPackard/galadriel/pkg/common/entity"
	"github.com/HewlettPackard/galadriel/pkg/server/api/admin"
	"github.com/HewlettPackard/galadriel/pkg/server/api/harvester"
	"github.com/HewlettPackard/galadriel/pkg/server/db/criteria"
)

type relationshipsParamGetter interface {
	GetPageSize() *int
	GetPageNumber() *int
	GetConsentStatus() *api.ConsentStatus
}

type adminParams struct {
	admin.GetRelationshipsParams
}

func (p adminParams) GetPageSize() *int {
	return p.PageSize
}

func (p adminParams) GetPageNumber() *int {
	return p.PageNumber
}

func (p adminParams) GetConsentStatus() *api.ConsentStatus {
	return p.ConsentStatus
}

type harvesterParams struct {
	harvester.GetRelationshipsParams
}

func (p harvesterParams) GetPageSize() *int {
	return p.PageSize
}

func (p harvesterParams) GetPageNumber() *int {
	return p.PageNumber
}

func (p harvesterParams) GetConsentStatus() *api.ConsentStatus {
	return p.ConsentStatus
}

func convertRelationshipsParamsToListCriteria(params relationshipsParamGetter) (*criteria.ListRelationshipsCriteria, error) {
	pageSize, err := convertPaginationParam(params.GetPageSize())
	if err != nil {
		return nil, err
	}

	pageNumber, err := convertPaginationParam(params.GetPageNumber())
	if err != nil {
		return nil, err
	}

	filterByConsentStatus, err := convertValidConsentStatusParam(params.GetConsentStatus())
	if err != nil {
		return nil, err
	}

	return &criteria.ListRelationshipsCriteria{
		FilterByConsentStatus: filterByConsentStatus,
		PageSize:              pageSize,
		PageNumber:            pageNumber,
		OrderByCreatedAt:      criteria.OrderDescending,
	}, nil
}

func convertPaginationParam(param *int) (uint, error) {
	if param == nil {
		return 0, nil
	}

	value := *param
	if value <= 0 {
		return 0, fmt.Errorf("pagination parameter must be a positive integer")
	}

	return uint(value), nil
}

func convertValidConsentStatusParam(status *api.ConsentStatus) (*entity.ConsentStatus, error) {
	if status == nil || *status == "" {
		return nil, nil
	}

	switch *status {
	case api.Approved, api.Denied, api.Pending:
		convertedStatus := entity.ConsentStatus(*status)
		return &convertedStatus, nil
	default:
		return nil, fmt.Errorf("invalid consent status filter %q, it must be one of [approved, denied, pending]", *status)
	}
}
