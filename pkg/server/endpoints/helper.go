package endpoints

import (
	"fmt"

	"github.com/HewlettPackard/galadriel/pkg/common/api"
	"github.com/HewlettPackard/galadriel/pkg/common/entity"
)

// QueryParamsAdapter is responsible for validating and adapting API values into
// valid params using business types
type QueryParamsAdapter struct {
	pageSize   *api.PageSize
	pageNumber *api.PageNumber

	consentStatus *api.ConsentStatus

	validParams ValidQueryParams
}

// ValidQueryParams is the struct that holds th valid types and validated values
// for query params
type ValidQueryParams struct {
	pageSize   uint
	pageNumber uint

	consentStatus *entity.ConsentStatus
}

// ValidateParams start the validation process over all query params
func (q *QueryParamsAdapter) ValidateParams() error {
	pageSize, pageNumber, err := q.validatePaginationParams(q.pageSize, q.pageNumber)
	if err != nil {
		return err
	}

	consentStatus, err := q.validateConsentStatusParam(q.consentStatus)
	if err != nil {
		return err
	}

	q.validParams = ValidQueryParams{
		pageSize:      pageSize,
		pageNumber:    pageNumber,
		consentStatus: consentStatus,
	}

	return nil
}

func (q *QueryParamsAdapter) validatePaginationParams(pgSize *int, pgNumber *int) (uint, uint, error) {
	pageSize := defaultPageSize
	pageNumber := defaultPageNumber

	if pgSize != nil {
		pageSize = *pgSize
		outOfLimits := pageSize < 0
		if outOfLimits {
			err := fmt.Errorf("page size %v is not accepted, must be positive", *pgSize)
			return 0, 0, err
		}
	}

	if pgNumber != nil {
		pageNumber = *pgNumber

		outOfLimits := pageNumber < 0
		if outOfLimits {
			err := fmt.Errorf("page number %v is not accepted, must be positive", *pgNumber)
			return 0, 0, err
		}
	}

	return uint(pageSize), uint(pageNumber), nil
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
