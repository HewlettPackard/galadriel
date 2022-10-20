package spire

import (
	"context"
	"errors"
	"fmt"
	"net"
	"testing"

	"github.com/HewlettPackard/galadriel/pkg/common/telemetry"
	"github.com/sirupsen/logrus"

	"github.com/spiffe/go-spiffe/v2/bundle/spiffebundle"
	"github.com/spiffe/go-spiffe/v2/spiffeid"
	"github.com/stretchr/testify/assert"
)

func TestNewLocalSpireServerSuccess(t *testing.T) {
	dialFn = func(ctx context.Context, addr net.Addr, makeClient clientMaker) (client, error) {
		return fakeInternalClient{}, nil
	}
	got := NewLocalSpireServer(context.Background(), &net.UnixAddr{})
	expected := &localSpireServer{
		client: fakeInternalClient{},
		logger: logrus.WithField(telemetry.SubsystemName, "local_spire_server"),
	}

	assert.NotNil(t, got)
	assert.Equal(t, expected, got)
}

func TestNewLocalSpireServerPanic(t *testing.T) {
	dialFn = func(ctx context.Context, addr net.Addr, makeClient clientMaker) (client, error) {
		return nil, errors.New("error from dial function")
	}
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("The function did not panic")
		}
	}()

	NewLocalSpireServer(context.Background(), &net.UnixAddr{})
}

func TestLocalSpireGetBundle(t *testing.T) {
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
			client := &fakeInternalClient{bundle: tt.expected}
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
