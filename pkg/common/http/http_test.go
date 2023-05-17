package http

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
)

type TestBody struct {
	Name string `json:"name"`
}

type HTTPSetup struct {
	EchoContext echo.Context
	Recorder    *httptest.ResponseRecorder
}

func Setup() *HTTPSetup {
	// Setup
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	return &HTTPSetup{
		Recorder:    rec,
		EchoContext: e.NewContext(req, rec),
	}
}

func TestWriteResponse(t *testing.T) {
	t.Run("Error when nil body is passed", func(t *testing.T) {
		setup := Setup()
		err := WriteResponse(setup.EchoContext, http.StatusOK, nil)
		assert.EqualError(t, err, "body is required")
		assert.Empty(t, setup.Recorder.Body)
	})

	t.Run("No error when an empty body is passed", func(t *testing.T) {
		setup := Setup()
		err := WriteResponse(setup.EchoContext, http.StatusOK, TestBody{})
		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, setup.Recorder.Code)
	})

	t.Run("Ensuring that the body is being full filled with the entity", func(t *testing.T) {
		expectedResponseBody := TestBody{Name: "teste"}
		setup := Setup()
		err := WriteResponse(setup.EchoContext, http.StatusOK, expectedResponseBody)
		assert.NoError(t, err)

		responseBody := TestBody{}

		err = json.Unmarshal(setup.Recorder.Body.Bytes(), &responseBody)
		assert.NoError(t, err)
		assert.Equal(t, expectedResponseBody, responseBody)
		assert.Equal(t, http.StatusOK, setup.Recorder.Code)
	})
}

func TestBodilessResponse(t *testing.T) {
	t.Run("Ensuring that the body is empty", func(t *testing.T) {
		setup := Setup()
		err := RespondWithoutBody(setup.EchoContext, http.StatusOK)
		assert.NoError(t, err)

		assert.NoError(t, err)
		assert.Empty(t, setup.Recorder.Body.Bytes())
		assert.Equal(t, http.StatusOK, setup.Recorder.Code)
	})
}

func TestFromBody(t *testing.T) {
	t.Run("Ensuring that the body is empty", func(t *testing.T) {
		setup := Setup()
		err := RespondWithoutBody(setup.EchoContext, http.StatusOK)
		assert.NoError(t, err)

		assert.NoError(t, err)
		assert.Empty(t, setup.Recorder.Body.Bytes())
		assert.Equal(t, http.StatusOK, setup.Recorder.Code)
	})
}
