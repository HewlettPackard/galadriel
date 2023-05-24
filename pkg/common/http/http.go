package http

import (
	"errors"
	"fmt"

	"github.com/labstack/echo/v4"
	"github.com/sirupsen/logrus"
)

// WriteResponse parses a struct into a json and writes in the response
func WriteResponse(ctx echo.Context, code int, body interface{}) error {
	if body == nil {
		return errors.New("body is required")
	}

	if err := ctx.JSON(code, body); err != nil {
		return fmt.Errorf("failed to write response body: %v", err)
	}

	return nil
}

// RespondWithoutBody wraps error echo bodiless responses.
func RespondWithoutBody(ctx echo.Context, code int) error {
	if err := ctx.NoContent(code); err != nil {
		return fmt.Errorf("failed to respond without body: %v", err)
	}

	return nil
}

// ParseRequestBodyToStruct parses json bytes into a struct
func ParseRequestBodyToStruct(ctx echo.Context, targetStruct interface{}) error {
	if targetStruct == nil {
		return fmt.Errorf("missing reference for handling parsed data")
	}

	return ctx.Bind(targetStruct)
}

// LogAndRespondWithError logs the error and returns an HTTP error.
func LogAndRespondWithError(logger logrus.FieldLogger, err error, errorMessage string, statusCode int) error {
	if err != nil {
		logger.Errorf("%s: %v", errorMessage, err)
	} else {
		logger.Error(errorMessage)
	}

	return &echo.HTTPError{
		Code:    statusCode,
		Message: errorMessage,
	}
}
