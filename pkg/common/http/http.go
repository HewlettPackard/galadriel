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
WriteObjectResponse parses a struct into a json and writes in the response
*/
func WriteObjectResponse[T any](ctx echo.Context, body *T) error {
	if body == nil {
		return nil
	}

	return write(ctx, body)
}

/*
WriteArrayResponse parses a slice of object into a json and writes in the response
*/
func WriteArrayResponse[T any](ctx echo.Context, body []*T) error {
	if body == nil {
		return nil
	}

	return write(ctx, body)
}

func write[T any](ctx echo.Context, body T) error {
	if err := ctx.JSON(http.StatusOK, body); err != nil {
		return fmt.Errorf("failed to write response body: %v", err)
	}
	return nil
}

/*
FromBody parses json bytes into a struct
*/
func FromBody[T any](ctx echo.Context, in *T) error {
	return ctx.Bind(in)
}
