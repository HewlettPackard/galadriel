package endpoints

import (
	"errors"
	"fmt"
	"time"

	"github.com/HewlettPackard/galadriel/pkg/common/api"
	"github.com/HewlettPackard/galadriel/pkg/common/entity"
	"github.com/deepmap/oapi-codegen/pkg/types"
)

func validatePaginationParams(pgSize *int, pgNumber *int) (uint, uint, error) {
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

func validateConsentStatusParam(status *api.ConsentStatus) (*entity.ConsentStatus, error) {
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

func validateTimeParams(startDate *types.Date, endDate *types.Date) (time.Time, time.Time, error) {
	from := defaultStartDate()
	until := defaultEndDate()

	if startDate != nil {
		if startDate.Time.After(until) {
			err := errors.New("can't use a startDate that is in the future")
			return time.Time{}, time.Time{}, err
		}

		from = startDate.Time
	}

	if endDate != nil {
		if endDate.Time.Before(until) && endDate.Time.After(from) {
			until = endDate.Time
		} else {
			err := errors.New("can't use a endDate that is before the startDate")
			return time.Time{}, time.Time{}, err
		}
	}

	if from.Add(30 * time.Minute).After(until) {
		err := errors.New("the minimal interval is 30 minutes")
		return time.Time{}, time.Time{}, err
	}

	return from, until, nil
}

func defaultStartDate() time.Time {
	// Last 30 Day
	return time.Now().Add(-30 * 24 * time.Hour)
}

func defaultEndDate() time.Time {
	return time.Now().Add(1 * time.Second)
}
