package http

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"

	"github.com/HewlettPackard/galadriel/cmd/common/cli"
)

const ()

type ErrorMessage struct {
	Message string
}

// NewUDSHTTPClient creates a new HTTP client configured to connect to a Unix
// Domain Socket (UDS) specified by the provided socketPath. The returned
// client uses custom transport that dials via the UDS, and is set to have
// a default timeout as specified in the cli.CommandTimeout.
func NewUDSHTTPClient(socketPath string) *http.Client {
	t := &http.Transport{
		DialContext: func(ctx context.Context, _, _ string) (net.Conn, error) {
			var d net.Dialer
			return d.DialContext(ctx, "unix", socketPath)
		},
	}

	return &http.Client{
		Transport: t,
		Timeout:   cli.CommandTimeout,
	}
}

// ReadResponse attempts to read the body of an HTTP response.
// It unmarshals any error message in the response body if the status code indicates an error.
// If the status code is not in the 2xx range, an error is returned.
func ReadResponse(res *http.Response) ([]byte, error) {
	if res == nil {
		return nil, fmt.Errorf("response is nil")
	}

	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	if res.StatusCode < http.StatusOK || res.StatusCode > http.StatusIMUsed {
		var errorMsg ErrorMessage
		if err := json.Unmarshal(body, &errorMsg); err != nil {
			return nil, fmt.Errorf("failed to unmarshal error message: %w", err)
		}

		return nil, fmt.Errorf("server returned status code %d with message: %s", res.StatusCode, errorMsg.Message)
	}

	return body, nil
}
