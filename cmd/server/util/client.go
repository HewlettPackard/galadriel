package util

import (
	"context"
	"io"
	"net"
	"net/http"
)

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
	r, err := c.client.Get("http://unix/token")
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
