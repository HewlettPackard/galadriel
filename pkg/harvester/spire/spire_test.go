package spire

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/HewlettPackard/galadriel/pkg/common"
	"github.com/spiffe/go-spiffe/v2/bundle/spiffebundle"
	"github.com/spiffe/go-spiffe/v2/spiffeid"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc"
)

func TestNewLocalSpireServerSuccess(t *testing.T) {
	// originalDialFn := &dialFn
	dialFn = func(ctx context.Context, path string, makeClient clientMaker) (client, error) {
		return fakeClient{}, nil
	}
	got := NewLocalSpireServer("")
	expected := &localSpireServer{
		client: fakeClient{},
		logger: *common.NewLogger("local_spire_server"),
	}

	assert.NotNil(t, got)
	assert.Equal(t, expected, got)
}

func TestNewLocalSpireServerPanic(t *testing.T) {
	dialFn = func(ctx context.Context, path string, makeClient clientMaker) (client, error) {
		return nil, errors.New("error from dial function")
	}
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("The function did not panic")
		}
	}()

	NewLocalSpireServer("")
}

func TestMakeSpireClient(t *testing.T) {
	got, err := makeSpireClient(&grpc.ClientConn{})
	assert.NotNil(t, got)
	assert.NoError(t, err)

	got, err = makeSpireClient(nil)
	assert.Nil(t, got)
	assert.EqualError(t, err, "grpc client connection is invalid")
}

func TestDialSocket(t *testing.T) {
	tests := []struct {
		name           string
		path           string
		expected       client
		err            string
		target         string
		clientErr      string
		dialContextErr string
	}{
		{
			name:     "ok_abs_path",
			path:     "/absolute/path",
			expected: &fakeClient{},
			target:   "unix:///absolute/path",
		}, {
			name:     "ok_rel_path",
			path:     "relative/path",
			expected: &fakeClient{},
			target:   "unix:relative/path",
		}, {
			name:           "dial_context_error",
			dialContextErr: "error_from_dial_context",
			err:            "failed to dial API socket: error_from_dial_context",
		},
		{
			name:      "make_client_error",
			clientErr: "error_from_make_client",
			err:       "failed to make client: error_from_make_client",
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			grpcDialContext = func(ctx context.Context, target string, opts ...grpc.DialOption) (*grpc.ClientConn, error) {
				if tt.dialContextErr != "" {
					return nil, errors.New(tt.dialContextErr)
				}
				if tt.path != "" {
					assert.Equal(t, tt.target, target)
				}
				return &grpc.ClientConn{}, nil
			}
			fakeMakeClient := func(conn *grpc.ClientConn) (client, error) {
				if tt.clientErr != "" {
					return nil, errors.New(tt.clientErr)
				}
				return tt.expected, nil
			}

			got, err := dialSocket(context.Background(), tt.path, fakeMakeClient)

			if tt.err != "" {
				assert.EqualError(t, err, tt.err)
				assert.Nil(t, got)
				return
			}

			assert.NoError(t, err)
			assert.Equal(t, tt.expected, got)
		})
	}

}

func TestGetBundle(t *testing.T) {
	tests := []struct {
		name     string
		expected *spiffebundle.Bundle
		err      string
	}{
		{
			name:     "ok",
			expected: spiffebundle.New(spiffeid.RequireTrustDomainFromString("example.org")),
		}, {
			name: "error_calling_client",
			err:  "error_from_client",
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			client := new(fakeClient)
			client.bundle = tt.expected
			if tt.err != "" {
				client.getBundleErr = errors.New(tt.err)
			}

			spire := &localSpireServer{client: client}

			got, err := spire.GetBundle(context.Background())
			if tt.err != "" {
				assert.EqualError(t, err, fmt.Sprintf("failed to get bundle: %v", tt.err))
				assert.Nil(t, got)
				return
			}

			assert.NoError(t, err)
			assert.Equal(t, tt.expected, got)
		})
	}
}
