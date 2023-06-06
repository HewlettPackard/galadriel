package endpoints

import (
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/HewlettPackard/galadriel/pkg/common/api"
	"github.com/HewlettPackard/galadriel/pkg/common/entity"
	chttp "github.com/HewlettPackard/galadriel/pkg/common/http"
	"github.com/deepmap/oapi-codegen/pkg/types"
	"github.com/labstack/echo/v4"
	"github.com/sirupsen/logrus"
)

func validatePaginationParams(
	echoCtx echo.Context,
	logger logrus.FieldLogger,
	pgSize *int,
	pgNumber *int,
) (uint, uint, error) {
	pageSize := defaultPageSize
	pageNumber := defaultPageNumber

	if pgSize != nil {
		pageSize = *pgSize
		outOfLimits := pageSize <= 0 || pageSize > 100
		if outOfLimits {
			errMsg := fmt.Errorf("page size %v is out of the accepted range [1 - 100]", *pgSize)
			return 0, 0, chttp.LogAndRespondWithError(logger, errMsg, errMsg.Error(), http.StatusBadRequest)
		}
	}

	if pgNumber != nil {
		pageNumber = *pgNumber
		// We may need to revisit the page number limitation in the future
		// This is just a first limitation to avoid crazy page numbers iex: pageNumber=100000
		outOfLimits := pageNumber < 0 || pageNumber > 100
		if outOfLimits {
			errMsg := fmt.Errorf("page number %v is out of the accepted range [0 - 100]", *pgSize)
			return 0, 0, chttp.LogAndRespondWithError(logger, errMsg, errMsg.Error(), http.StatusBadRequest)
		}
	}

	return uint(pageSize), uint(pageNumber), nil
}

func validateConsentStatusParam(
	echoCtx echo.Context,
	logger logrus.FieldLogger,
	status *api.ConsentStatus,
) (*entity.ConsentStatus, error) {
	if status != nil {
		switch *status {
		case api.Approved, api.Denied, api.Pending:
		default:
			err := fmt.Errorf("status filter %q is not supported, available values [approved, denied, pending]", *status)
			return nil, chttp.LogAndRespondWithError(logger, err, err.Error(), http.StatusBadRequest)
		}

		consentStatus := entity.ConsentStatus(*status)

		return &consentStatus, nil
	}

	return nil, nil
}

func validateTimeParams(
	echoCtx echo.Context,
	logger logrus.FieldLogger,
	startDate *types.Date,
	endDate *types.Date,
) (time.Time, time.Time, error) {
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
