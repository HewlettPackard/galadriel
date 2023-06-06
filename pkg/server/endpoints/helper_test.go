package endpoints

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/HewlettPackard/galadriel/pkg/common/api"
	"github.com/HewlettPackard/galadriel/pkg/common/entity"
	"github.com/deepmap/oapi-codegen/pkg/types"
	"github.com/labstack/echo/v4"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

func NewEchoContext(t *testing.T) echo.Context {
	e := echo.New()
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "http://fakeurl", nil)
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	echoCtx := e.NewContext(req, rec)
	return echoCtx

}
func TestValidatePaginationParams(t *testing.T) {
	testCases := []struct {
		name          string
		pageSize      int
		pageNumber    int
		noParams      bool
		expectedError error
	}{
		{
			name:          "Page size out of range",
			pageSize:      101,
			pageNumber:    0,
			expectedError: errors.New("code=400, message=page size 101 is out of the accepted range [1 - 100]"),
		},
		{
			name:          "Page number out of range",
			pageSize:      10,
			pageNumber:    101,
			expectedError: errors.New("code=400, message=page number 10 is out of the accepted range [0 - 100]"),
		},
		{
			name:       "Successfully pass verifications",
			pageSize:   10,
			pageNumber: 1,
		},
		{
			name:     "Nil pagination params",
			noParams: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			echoCtx := NewEchoContext(t)

			var err error
			var pageSize, pageNumber uint
			var logger logrus.FieldLogger = logrus.StandardLogger().WithField("test", tc.name)

			if tc.noParams {
				pageSize, pageNumber, err = validatePaginationParams(echoCtx, logger, nil, nil)
			} else {
				pageSize, pageNumber, err = validatePaginationParams(echoCtx, logger, &tc.pageSize, &tc.pageNumber)
			}

			if tc.expectedError != nil {
				assert.Zero(t, pageSize)
				assert.Zero(t, pageNumber)
				assert.EqualError(t, err, tc.expectedError.Error())
			} else {
				if tc.noParams {
					assert.NoError(t, err)
					assert.Equal(t, uint(defaultPageSize), pageSize)
					assert.Equal(t, uint(defaultPageNumber), pageNumber)
				} else {
					assert.NoError(t, err)
					assert.Equal(t, uint(tc.pageSize), pageSize)
					assert.Equal(t, uint(tc.pageNumber), pageNumber)

				}
			}
		})
	}
}

func TestValidateConsentStatusParam(t *testing.T) {
	testCases := []struct {
		name                string
		expectedError       error
		noParams            bool
		consentStatus       api.ConsentStatus
		entityConsentStatus entity.ConsentStatus
	}{
		{
			name:                "Approved",
			consentStatus:       api.Approved,
			entityConsentStatus: entity.ConsentStatusApproved,
		},
		{
			name:                "Denied",
			consentStatus:       api.Denied,
			entityConsentStatus: entity.ConsentStatusDenied,
		},
		{
			name:                "Pending",
			consentStatus:       api.Pending,
			entityConsentStatus: entity.ConsentStatusPending,
		},
		{
			name:     "Nil filter should pass verification",
			noParams: true,
		},
		{
			name:          "Unsuported filter type",
			consentStatus: "teste",
			expectedError: errors.New("code=400, message=status filter \"teste\" is not supported, available values [approved, denied, pending]"),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			echoCtx := NewEchoContext(t)

			var consentStatus *api.ConsentStatus
			if !tc.noParams {
				consentStatus = &tc.consentStatus
			}

			var logger logrus.FieldLogger = logrus.StandardLogger().WithField("test", tc.name)
			status, err := validateConsentStatusParam(echoCtx, logger, consentStatus)

			if tc.expectedError != nil {
				assert.Empty(t, status)
				assert.EqualError(t, err, tc.expectedError.Error())
			} else {
				if tc.noParams {
					assert.Nil(t, status)
				} else {
					assert.NoError(t, err)
					assert.Equal(t, string(*status), string(tc.consentStatus))
				}
			}
		})
	}
}

func TestValidateTimeParams(t *testing.T) {
	testCases := []struct {
		name              string
		startDate         *types.Date
		endDate           *types.Date
		expectErr         error
		expectedEndDate   time.Time
		expectedStartDate time.Time
	}{
		{
			name:              "Applying Default values",
			expectedEndDate:   defaultEndDate(),
			expectedStartDate: defaultStartDate(),
		},
		{
			name:              "Successfully pass verification",
			startDate:         &types.Date{time.Now().Add(-2 * time.Hour)},
			endDate:           &types.Date{time.Now().Add(-1 * time.Hour)},
			expectedEndDate:   time.Now().Add(-1 * time.Hour),
			expectedStartDate: time.Now().Add(-2 * time.Hour),
		},
		{
			name:      "StartDate that is in the future",
			startDate: &types.Date{time.Now().Add(1 * time.Minute)},
			endDate:   &types.Date{time.Now()},
			expectErr: errors.New("can't use a startDate that is in the future"),
		},
		{
			name:      "EndDate before startDate",
			startDate: &types.Date{time.Now().Add(-1 * time.Hour)},
			endDate:   &types.Date{time.Now().Add(-2 * time.Hour)},
			expectErr: errors.New("can't use a endDate that is before the startDate"),
		},
		{
			name:      "Checking at least 30 minutes interval",
			startDate: &types.Date{time.Now().Add(-1 * time.Hour)},
			endDate:   &types.Date{time.Now().Add(-45 * time.Minute)},
			expectErr: errors.New("the minimal interval is 30 minutes"),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			echoCtx := NewEchoContext(t)

			var logger logrus.FieldLogger = logrus.StandardLogger().WithField("test", tc.name)
			startDate, endDate, err := validateTimeParams(echoCtx, logger, tc.startDate, tc.endDate)

			if tc.expectErr != nil {
				assert.Empty(t, endDate)
				assert.Empty(t, startDate)
				assert.EqualError(t, err, tc.expectErr.Error())
			} else {
				assert.NoError(t, err)
				assertTime(t, tc.expectedEndDate, endDate)
				assertTime(t, tc.expectedStartDate, startDate)
			}
		})
	}
}

// Compare 2 times ignoring the milliseconds for testing running purpose
func assertTime(t *testing.T, expectedTime time.Time, actual time.Time) {
	assert.Equal(t, expectedTime.Year(), actual.Year())
	assert.Equal(t, expectedTime.Month(), actual.Month())
	assert.Equal(t, expectedTime.Day(), actual.Day())
	assert.Equal(t, expectedTime.Hour(), actual.Hour())
	assert.Equal(t, expectedTime.Minute(), actual.Minute())
	assert.Equal(t, expectedTime.Second(), actual.Second())
}
