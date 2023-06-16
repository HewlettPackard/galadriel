package endpoints

import (
	"testing"

	"github.com/HewlettPackard/galadriel/pkg/common/api"
	"github.com/HewlettPackard/galadriel/pkg/common/entity"
	"github.com/stretchr/testify/assert"
)

type mockParams struct {
	pageSize      *int
	pageNumber    *int
	consentStatus *api.ConsentStatus
}

func (m *mockParams) GetPageSize() *int {
	return m.pageSize
}

func (m *mockParams) GetPageNumber() *int {
	return m.pageNumber
}

func (m *mockParams) GetConsentStatus() *api.ConsentStatus {
	return m.consentStatus
}

func TestConvertRelationshipsParamsToListCriteria(t *testing.T) {
	ps := 1
	pn := 1
	cs := api.Approved

	params := &mockParams{
		pageSize:      &ps,
		pageNumber:    &pn,
		consentStatus: &cs,
	}

	result, err := convertRelationshipsParamsToListCriteria(params)
	assert.Nil(t, err)

	assert.Equal(t, uint(ps), result.PageSize)
	assert.Equal(t, uint(pn), result.PageNumber)
	assert.Equal(t, entity.ConsentStatus(cs), *result.FilterByConsentStatus)
}

func TestConvertPaginationParam(t *testing.T) {
	value := 1

	result, err := convertPaginationParam(&value)
	assert.Nil(t, err)
	assert.Equal(t, uint(value), result)

	value = -1
	result, err = convertPaginationParam(&value)
	assert.NotNil(t, err)
	assert.Equal(t, uint(0), result)
}

func TestConvertValidConsentStatusParam(t *testing.T) {
	cs := api.Approved

	result, err := convertValidConsentStatusParam(&cs)
	assert.Nil(t, err)
	assert.Equal(t, entity.ConsentStatus(cs), *result)

	cs = "invalid"
	result, err = convertValidConsentStatusParam(&cs)
	assert.NotNil(t, err)
	assert.Nil(t, result)
}
