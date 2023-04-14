package http

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/labstack/echo/v4"
)

/*
HandleError default error handler function over the application
*/
func HandleError(w http.ResponseWriter, errMsg string) error {
	errBytes := []byte(errMsg)
	w.WriteHeader(500)
	_, err := w.Write(errBytes)
	if err != nil {
		return fmt.Errorf("failed to write error response: %v", err)
	}

	encoder := json.NewEncoder(w)
	encoder.SetEscapeHTML(true)
	err = encoder.Encode(errBytes)
	if err != nil {
		return fmt.Errorf("failed to write error response: %v", err)
	}

	return nil
}

func HandleTCPError(ctx echo.Context, err error) error {
	_, err = ctx.Response().Write([]byte(err.Error()))
	if err != nil {
		return fmt.Errorf("failed to write error response: %v", err)
	}

	return nil
}

/*
WriteAsJSInResponse parses a struct into a json and writes in the response
w: The response writer to be full filled with the struct response bytes
out: A pointer to the interface to be writed in the response
*/
func WriteAsJSInResponse(w http.ResponseWriter, out interface{}) error {
	if out == nil {
		return nil
	}

	outBytes, err := json.Marshal(out)
	if err != nil {
		return fmt.Errorf("failed marshal response body: %v", err)
	}

	_, err = w.Write(outBytes)
	if err != nil {
		return fmt.Errorf("failed to write response body: %v", err)
	}

	return nil
}

// Writes the response to the client
func WriteResponse(ctx echo.Context, body interface{}) error {
	if body == nil {
		return nil
	}

	if err := ctx.JSON(http.StatusOK, body); err != nil {
		return fmt.Errorf("failed to write response body: %v", err)
	}

	return nil
}

/*
FromJSBody parses json bytes into a struct
r: Request that contains the json bytes to be parsed into 'in'
in: Reference(pointer) to the interface to be full filled
*/
func FromJSBody(r *http.Request, in interface{}) error {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		return fmt.Errorf("failed reading request body: %v", err)
	}

	if err = json.Unmarshal(body, in); err != nil {
		return fmt.Errorf("failed unmarshalling request: %v", err)
	}

	return nil
}

func FromBody(ctx echo.Context, in interface{}) error {
	return ctx.Bind(in)
}
