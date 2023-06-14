package endpoints

import (
	"fmt"

	"github.com/HewlettPackard/galadriel/pkg/common/api"
	"github.com/HewlettPackard/galadriel/pkg/common/entity"
	"github.com/HewlettPackard/galadriel/pkg/server/db/criteria"
)

// QueryParamsAdapter is responsible for validating and adapting API values into
// valid params using business types
type QueryParamsAdapter struct {
	pageSize   *api.PageSize
	pageNumber *api.PageNumber

	consentStatus *api.ConsentStatus

	validParams ValidQueryParams

	order *api.CreatedAtOrder
}

// ValidQueryParams is the struct that holds th valid types and validated values
// for query params
type ValidQueryParams struct {
	pageSize      uint
	pageNumber    uint
	order         criteria.OrderDirection
	consentStatus *entity.ConsentStatus
}

// ValidateParams start the validation process over all query params
func (q *QueryParamsAdapter) ValidateParams() error {
	pageSize, pageNumber, err := q.validatePaginationParams(q.pageSize, q.pageNumber)
	if err != nil {
		return err
	}

	order, err := q.validateOrder(q.order)
	if err != nil {
		return err
	}

	consentStatus, err := q.validateConsentStatusParam(q.consentStatus)
	if err != nil {
		return err
	}

	q.validParams = ValidQueryParams{
		order:         *order,
		pageSize:      pageSize,
		pageNumber:    pageNumber,
		consentStatus: consentStatus,
	}

	return nil
}

func (q *QueryParamsAdapter) validatePaginationParams(pgSize *int, pgNumber *int) (uint, uint, error) {
	pageSize := uint(0)
	pageNumber := uint(0)

	if pgSize != nil {
		pageSize = uint(*pgSize)
		if *pgSize < 0 {
			err := fmt.Errorf("page size %v is not accepted, must be positive", *pgSize)
			return 0, 0, err
		}
	}

	if pgNumber != nil {
		pageNumber = uint(*pgNumber)
		if *pgNumber < 0 {
			err := fmt.Errorf("page number %v is not accepted, must be positive", *pgNumber)
			return 0, 0, err
		}
	}

	return pageSize, pageNumber, nil
}

func (q *QueryParamsAdapter) validateConsentStatusParam(status *api.ConsentStatus) (*entity.ConsentStatus, error) {
	if status != nil {
		switch *status {
		case api.Approved, api.Denied, api.Pending:
		default:
			err := fmt.Errorf("status filter %q is not supported, available values [approved, denied, pending]", *status)
			return nil, err
		}

		consentStatus := entity.ConsentStatus(*status)
		return &consentStatus, nil
	}

	return nil, nil
}

func (q *QueryParamsAdapter) validateOrder(apiOrder *api.CreatedAtOrder) (*criteria.OrderDirection, error) {
	order := criteria.NoOrder
	if apiOrder != nil {
		order = criteria.OrderDirection(*apiOrder)
		switch order {
		case criteria.NoOrder, criteria.OrderAscending, criteria.OrderDescending:
		default:
			err := fmt.Errorf("order filter %q is not supported, available values [asc, desc]", order)
			return nil, err
		}
	}

	return &order, nil
}
