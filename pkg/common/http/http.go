package http

import (
	"fmt"
	"net/http"

	"github.com/labstack/echo/v4"
)

/*
NewHTTPErr normalize the error to be handled by echo erro handler
*/
func NewHTTPErr(ctx echo.Context, err error, code int) error {
	return echo.NewHTTPError(code, err.Error())
}

/*
WriteResponse parses a struct into a json and writes in the response
*/
func WriteResponse(ctx echo.Context, body interface{}) error {
	if body == nil {
		return fmt.Errorf("missing body for writing in response")
	}

	if err := ctx.JSON(http.StatusOK, body); err != nil {
		return fmt.Errorf("failed to write response body: %v", err)
	}

	return nil
}

/*
FromBody parses json bytes into a struct
*/
func FromBody(ctx echo.Context, in interface{}) error {
	if in == nil {
		return fmt.Errorf("missing reference for handling parsed data")
	}

	return ctx.Bind(in)
}
