package http

import (
	"fmt"
	"net/http"

	"github.com/labstack/echo/v4"
)

/*
HandleError default error handler function over the application
*/
func HandleTCPError(ctx echo.Context, err error) error {
	_, err = ctx.Response().Write([]byte(err.Error()))
	if err != nil {
		return fmt.Errorf("failed to write error response: %v", err)
	}

	return nil
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
