package http

import (
	"errors"
	"fmt"

	"github.com/labstack/echo/v4"
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

// FromBody parses json bytes into a struct
func FromBody(ctx echo.Context, in interface{}) error {
	if in == nil {
		return fmt.Errorf("missing reference for handling parsed data")
	}

	return ctx.Bind(in)
}
