package http

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/labstack/echo/v4"
)

// WriteResponse parses a struct into a json and writes in the response
func WriteResponse(ctx echo.Context, body interface{}) error {
	if body == nil {
		return errors.New("body is required")
	}

	if err := ctx.JSON(http.StatusOK, body); err != nil {
		return fmt.Errorf("failed to write response body: %v", err)
	}

	return nil
}

// BodylessResponse wraps error echo body-less responses.
func BodylessResponse(ctx echo.Context) error {
	if err := ctx.NoContent(http.StatusOK); err != nil {
		return fmt.Errorf("failed to respond without body: %v", err)
	}

	return nil
}

// FromBody parses json bytes into a struct
func FromBody(ctx echo.Context, in interface{}) error {
	if in == nil {
		return fmt.Errorf("missing reference for handling parsed data")
	}

	return ctx.Bind(in)
}
