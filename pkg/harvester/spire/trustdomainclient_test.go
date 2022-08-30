package spire

import (
	"context"
	"crypto/x509"
	"errors"
	"fmt"
	"testing"

	"github.com/spiffe/go-spiffe/v2/bundle/spiffebundle"
	"github.com/spiffe/go-spiffe/v2/spiffeid"
	"github.com/spiffe/spire-api-sdk/proto/spire/api/types"
	"github.com/stretchr/testify/assert"
)

func makeTypeFederationRelationship(td string, profile BundleEndpointProfile) *types.FederationRelationship {
	if td == "" {
		return &types.FederationRelationship{}
	}
	out := &types.FederationRelationship{
		TrustDomain:       td,
		TrustDomainBundle: &types.Bundle{TrustDomain: td},
		BundleEndpointUrl: fmt.Sprintf("https://%s/bundle", td),
	}
	switch any(profile).(type) {
	case *HTTPSSpiffeBundleEndpointProfile:
		out.BundleEndpointProfile = &types.FederationRelationship_HttpsSpiffe{
			HttpsSpiffe: &types.HTTPSSPIFFEProfile{
				EndpointSpiffeId: fmt.Sprintf("spiffe://%s/spire/server", td),
			},
		}
	case *HTTPSWebBundleEndpointProfile:
		out.BundleEndpointProfile = &types.FederationRelationship_HttpsWeb{
			HttpsWeb: &types.HTTPSWebProfile{},
		}
	default:
		panic("unsupported Bundle Endpoint Profile")
	}

	return out

}

func makeTypeFederationRelationships(number int) []*types.FederationRelationship {
	var out []*types.FederationRelationship

	for i := 0; i < number; i++ {
		td := fmt.Sprintf("%d.org", i)
		out = append(out, makeTypeFederationRelationship(td, &HTTPSSpiffeBundleEndpointProfile{}))
	}

	return out
}

func makeFederationRelationshipResult(td string) *FederationRelationship {
	trustDomain := spiffeid.RequireTrustDomainFromString(td)
	out := &FederationRelationship{
		TrustDomain:       trustDomain,
		TrustDomainBundle: spiffebundle.New(trustDomain),
		BundleEndpointURL: fmt.Sprintf("https://%s/bundle", td),
		BundleEndpointProfile: HTTPSSpiffeBundleEndpointProfile{
			SpiffeID: spiffeid.RequireFromStringf("spiffe://%s/spire/server", td),
		},
	}
	out.TrustDomainBundle.SetX509Authorities([]*x509.Certificate{})

	return out
}

func makeFederationRelationshipResults(number int) []*FederationRelationship {
	var out []*FederationRelationship

	for i := 0; i < number; i++ {
		td := fmt.Sprintf("%d.org", i)
		out = append(out, makeFederationRelationshipResult(td))
	}

	return out
}

func TestNewTrustDomainClientSuccess(t *testing.T) {
	got := NewTrustDomainClient(fakeClientConn{})

	assert.NotNil(t, got)
	assert.IsType(t, trustDomainClient{}, got)
}

func TestClientListFederationRelationships(t *testing.T) {
	tests := []struct {
		name                   string
		expected               []*FederationRelationship
		federatedRelationships []*types.FederationRelationship
		pageSize               int
		err                    string
		clientErr              string
	}{
		{
			name:                   "ok",
			expected:               makeFederationRelationshipResults(10),
			federatedRelationships: makeTypeFederationRelationships(10),
		}, {
			name:                   "ok_small_pagination",
			expected:               makeFederationRelationshipResults(10),
			federatedRelationships: makeTypeFederationRelationships(10),
			pageSize:               3,
		}, {
			name:                   "ok_pagination_overflow",
			expected:               makeFederationRelationshipResults(2),
			federatedRelationships: makeTypeFederationRelationships(2),
			pageSize:               3,
		}, {
			name:                   "ok_pagination",
			expected:               makeFederationRelationshipResults(3),
			federatedRelationships: makeTypeFederationRelationships(3),
			pageSize:               3,
		},
		{
			name:      "client_error",
			clientErr: "error_from_client",
			err:       "failed to list federation relationships: error_from_client",
		},
		{
			name:                   "error_parsing_proto",
			federatedRelationships: []*types.FederationRelationship{{}},
			err:                    "failed to parse federation relationships: failed to parse federated relationship: failed to parse federated trust domain: trust domain is missing",
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			spireTrustDomainClient := &fakeSpireTrustDomainClient{}
			spireTrustDomainClient.federationRelationships = tt.federatedRelationships
			if tt.pageSize != 0 {

				listFederationRelationshipsPageSize = tt.pageSize
			}

			client := &trustDomainClient{client: spireTrustDomainClient}

			if tt.clientErr != "" {
				spireTrustDomainClient.batchListFederationRelationshipsError = errors.New(tt.clientErr)
			}

			got, err := client.ListFederationRelationships(context.Background())

			if tt.err != "" {
				assert.EqualError(t, err, tt.err)
				assert.Nil(t, got)
				return
			}

			assert.Equal(t, tt.expected, got)
			assert.Nil(t, err)
		})
	}
}
