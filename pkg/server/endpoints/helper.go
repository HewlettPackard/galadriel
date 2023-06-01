package endpoints

import (
	"fmt"
	"net/http"

	"github.com/HewlettPackard/galadriel/pkg/common/api"
	"github.com/HewlettPackard/galadriel/pkg/common/entity"
	chttp "github.com/HewlettPackard/galadriel/pkg/common/http"
	"github.com/labstack/echo/v4"
	"github.com/sirupsen/logrus"
)

func validatePaginationParams(
	echoCtx echo.Context,
	logger logrus.FieldLogger,
	pgSize *int,
	pgNumber *int,
) (int, int, error) {
	pageSize := defaultPageSize
	pageNumber := defaultPageNumber

	if pgSize != nil && (*pgSize < 0 || *pgSize > 100) {
		errMsg := fmt.Errorf("page size %s is out of the accepted range [1 - 100]", *pgSize)
		return 0, 0, chttp.LogAndRespondWithError(logger, errMsg, errMsg.Error(), http.StatusBadRequest)
	}

	// We may need to revisit the page number limitation in the future
	// This is just a first limitation to avoid crazy page numbers
	if pgNumber != nil && (*pgNumber < 0 || *pgNumber > 100) {
		errMsg := fmt.Errorf("page number %s is out of the accepted range [0 - 100]", *pgSize)
		return 0, 0, chttp.LogAndRespondWithError(logger, errMsg, errMsg.Error(), http.StatusBadRequest)
	}

	return pageSize, pageNumber, nil
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
			err := fmt.Errorf("status filter %q is not supported, approved values [approved, denied, pending]", *status)
			return nil, chttp.LogAndRespondWithError(logger, err, err.Error(), http.StatusBadRequest)
		}

		consentStatus := entity.ConsentStatus(*status)

		return &consentStatus, nil
	}

	return nil, nil
}
