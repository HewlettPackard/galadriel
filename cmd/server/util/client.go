package util

import (
	"context"
	"fmt"
	"io"
	"net"
	"net/http"
)

// URL pattern to make http calls on local Unix domain socket,
// the Host is required for the URL, but it's not relevant
const localURL = "http://local/%s"
const tokenPath = "token"

// ServerLocalClient represents a local client of the Galadriel Server.
type ServerLocalClient interface {
	GenerateJoinToken() (string, error)
}

// TODO: improve this adding options for the transport, dialcontext, and http.Client.
func NewServerClient(socketPath string) ServerLocalClient {
	t := &http.Transport{
		DialContext: func(ctx context.Context, _, _ string) (net.Conn, error) {
			var d net.Dialer
			return d.DialContext(ctx, "unix", socketPath)
		}}
	c := &http.Client{
		Transport: t,
	}

	return serverClient{client: c}
}

type serverClient struct {
	client *http.Client
}

func (c serverClient) GenerateJoinToken() (string, error) {
	tokenURL := fmt.Sprintf(localURL, tokenPath)
	r, err := c.client.Get(tokenURL)
	if err != nil {
		return "", err
	}
	defer r.Body.Close()

	b, err := io.ReadAll(r.Body)
	if err != nil {
		return "", err
	}
	return string(b), nil
}
