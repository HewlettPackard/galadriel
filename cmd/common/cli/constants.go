package cli

import "time"

const (
	// LocalhostURL is set to "http://localhost/" because all API calls
	// are routed through a Unix Domain Socket (UDS) and not over
	// a conventional network. The HTTP client ignores the actual
	// address and sends all requests to the local UDS path. This
	// is just a placeholder URL to satisfy the http.Client's
	// requirement for a valid URL format.
	LocalhostURL = "http://localhost/"

	CommandTimeout = 5 * time.Second
)
