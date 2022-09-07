package spire

import (
	"context"
	"crypto/x509"
	"errors"
	"testing"

	"github.com/spiffe/go-spiffe/v2/bundle/spiffebundle"
	"github.com/spiffe/go-spiffe/v2/spiffeid"
	"github.com/spiffe/spire-api-sdk/proto/spire/api/types"
	"github.com/stretchr/testify/assert"
)

func TestNewBundleClientSuccess(t *testing.T) {
	got := NewBundleClient(fakeClientConn{})

	assert.NotNil(t, got)
	assert.IsType(t, bundleClient{}, got)
}

func TestClientGetBundle(t *testing.T) {
	tests := []struct {
		name          string
		expected      *spiffebundle.Bundle
		err           string
		clientErr     string
		clientBundle  *types.Bundle
		protoParseErr string
	}{
		{
			name:         "ok",
			expected:     spiffebundle.New(spiffeid.RequireTrustDomainFromString("example.org")),
			clientBundle: &types.Bundle{TrustDomain: "example.org"},
		},
		{
			name:      "error_calling_client",
			err:       "failed to get bundle from trust domain client: error_from_client",
			clientErr: "error_from_client",
		}, {
			name:          "error_parsing_client_response",
			clientBundle:  &types.Bundle{},
			err:           "failed to parse spire server bundle response: failed to parse trust domain: trust domain is missing",
			protoParseErr: "failed to parse trust domain: trust domain is missing",
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			spireBundleClient := &fakeSpireBundleClient{bundle: tt.clientBundle}
			if tt.expected != nil {
				tt.expected.SetX509Authorities([]*x509.Certificate{}) // TODO: check this
			}
			if tt.clientErr != "" {
				spireBundleClient.getBundleErr = errors.New(tt.clientErr)
			}

			client := &bundleClient{client: spireBundleClient}

			got, err := client.GetBundle(context.Background())

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
