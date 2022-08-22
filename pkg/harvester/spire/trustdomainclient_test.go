package spire

import (
	"context"
	"crypto/x509"
	"errors"
	"testing"

	"github.com/spiffe/go-spiffe/v2/bundle/spiffebundle"
	"github.com/spiffe/go-spiffe/v2/spiffeid"
	trustdomainv1 "github.com/spiffe/spire-api-sdk/proto/spire/api/server/trustdomain/v1"
	"github.com/spiffe/spire-api-sdk/proto/spire/api/types"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc/codes"
)

var (
	inValidHttpsSpiffeExample = &FederationRelationship{
		TrustDomain: spiffeid.RequireTrustDomainFromString("example.org"),
		BundleEndpointProfile: HTTPSSpiffeBundleEndpointProfile{
			SpiffeID: spiffeid.RequireFromString("spiffe://example.org/spire/server"),
		},
		TrustDomainBundle: spiffebundle.New(spiffeid.RequireTrustDomainFromString("example.org")),
		BundleEndpointURL: "https://example.org/bundle",
	}
	outValidHttpsSpiffeExample = &FederationRelationshipResult{
		FederationRelationship: &FederationRelationship{
			TrustDomain: spiffeid.RequireTrustDomainFromString("example.org"),
			BundleEndpointProfile: HTTPSSpiffeBundleEndpointProfile{
				SpiffeID: spiffeid.RequireFromString("spiffe://example.org/spire/server"),
			},
			TrustDomainBundle: spiffebundle.New(spiffeid.RequireTrustDomainFromString("example.org")),
			BundleEndpointURL: "https://example.org/bundle",
		},
		Status: &FederationRelationshipResultStatus{Code: codes.OK},
	}
	outValidHttpsWebExample = &FederationRelationshipResult{
		FederationRelationship: &FederationRelationship{
			TrustDomain:           spiffeid.RequireTrustDomainFromString("example.org"),
			BundleEndpointProfile: HTTPSWebBundleEndpointProfile{},
			TrustDomainBundle:     spiffebundle.New(spiffeid.RequireTrustDomainFromString("example.org")),
			BundleEndpointURL:     "https://example.org/bundle",
		},
		Status: &FederationRelationshipResultStatus{Code: codes.OK},
	}
	clientOkHttpsSpiffeExample = &trustdomainv1.BatchCreateFederationRelationshipResponse_Result{
		Status: &types.Status{Code: int32(codes.OK)},
		FederationRelationship: &types.FederationRelationship{
			TrustDomain: "example.org",
			BundleEndpointProfile: &types.FederationRelationship_HttpsSpiffe{
				HttpsSpiffe: &types.HTTPSSPIFFEProfile{
					EndpointSpiffeId: "spiffe://example.org/spire/server",
				},
			},
			TrustDomainBundle: &types.Bundle{TrustDomain: "example.org"},
			BundleEndpointUrl: "https://example.org/bundle",
		},
	}
	clientOkHttpsWebExample = &trustdomainv1.BatchCreateFederationRelationshipResponse_Result{
		Status: &types.Status{Code: int32(codes.OK)},
		FederationRelationship: &types.FederationRelationship{
			TrustDomain: "example.org",
			BundleEndpointProfile: &types.FederationRelationship_HttpsWeb{
				HttpsWeb: &types.HTTPSWebProfile{},
			},
			TrustDomainBundle: &types.Bundle{TrustDomain: "example.org"},
			BundleEndpointUrl: "https://example.org/bundle",
		},
	}
	clientInvalidHttpsWebExample = &trustdomainv1.BatchCreateFederationRelationshipResponse_Result{
		Status:                 &types.Status{Code: int32(codes.OK)},
		FederationRelationship: &types.FederationRelationship{
			// TrustDomain: "example.org",
			// BundleEndpointProfile: &types.FederationRelationship_HttpsWeb{
			// 	HttpsWeb: &types.HTTPSWebProfile{},
			// },
			// TrustDomainBundle: &types.Bundle{TrustDomain: "example.org"},
			// BundleEndpointUrl: "https://example.org/bundle",
		},
	}
)

func TestNewTrustDomainClientSuccess(t *testing.T) {
	got := NewTrustDomainClient(fakeClientConn{})

	assert.NotNil(t, got)
	assert.IsType(t, trustDomainClient{}, got)
}

func TestClientCreateFederationRelationships(t *testing.T) {
	tests := []struct {
		name                    string
		expected                []*FederationRelationshipResult
		input                   []*FederationRelationship
		federationRelationships []*FederationRelationship
		err                     string
		clientResponse          *trustdomainv1.BatchCreateFederationRelationshipResponse
		clientErr               string
	}{
		{
			name:     "ok_https",
			input:    []*FederationRelationship{inValidHttpsSpiffeExample},
			expected: []*FederationRelationshipResult{outValidHttpsSpiffeExample},
			clientResponse: &trustdomainv1.BatchCreateFederationRelationshipResponse{
				Results: []*trustdomainv1.BatchCreateFederationRelationshipResponse_Result{clientOkHttpsSpiffeExample},
			},
		}, {
			name:     "ok_web",
			input:    []*FederationRelationship{inValidHttpsSpiffeExample},
			expected: []*FederationRelationshipResult{outValidHttpsWebExample},
			clientResponse: &trustdomainv1.BatchCreateFederationRelationshipResponse{
				Results: []*trustdomainv1.BatchCreateFederationRelationshipResponse_Result{clientOkHttpsWebExample},
			},
		},
		{
			name:  "error_parsing_proto",
			input: []*FederationRelationship{{TrustDomain: spiffeid.TrustDomain{}}},
			err:   "failed to convert federation relationships to proto: failed to convert trust domain bundle to proto: trust domain bundle must be set",
		},
		{
			name:      "client_error",
			input:     []*FederationRelationship{inValidHttpsSpiffeExample},
			clientErr: "error_from_client",
			err:       "failed to create federation relationships: error_from_client",
		},
		{
			name:  "invalid_client_response",
			input: []*FederationRelationship{inValidHttpsSpiffeExample},
			clientResponse: &trustdomainv1.BatchCreateFederationRelationshipResponse{
				Results: []*trustdomainv1.BatchCreateFederationRelationshipResponse_Result{clientInvalidHttpsWebExample},
			},
			err: "failed to parse federation relationship results: failed to convert federation relationship to proto: failed to parse federated trust domain: trust domain is missing",
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			spireTrustDomainClient := &fakeSpireTrustDomainClient{}
			if tt.expected != nil {
				for _, r := range tt.expected {
					r.FederationRelationship.TrustDomainBundle.SetX509Authorities([]*x509.Certificate{})
				}
			}

			spireTrustDomainClient.batchCreateFederationRelationshipResponse = tt.clientResponse
			if tt.clientErr != "" {
				spireTrustDomainClient.batchCreateFederationRelationshipError = errors.New(tt.clientErr)
			}

			client := &trustDomainClient{client: spireTrustDomainClient}

			got, err := client.CreateFederationRelationships(context.Background(), tt.input)

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
