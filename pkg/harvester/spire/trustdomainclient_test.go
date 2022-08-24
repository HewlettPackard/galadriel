package spire

import (
	"context"
	"crypto/x509"
	"errors"
	"fmt"
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
	outDeleteStatusExample = &FederationRelationshipResult{
		Status: &FederationRelationshipResultStatus{
			Code:        codes.OK,
			TrustDomain: "example.org",
		},
	}
)

func makeFederationRelationship(td string, profile BundleEndpointProfile) *types.FederationRelationship {
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

func TestNewTrustDomainClientSuccess(t *testing.T) {
	got := NewTrustDomainClient(fakeClientConn{})

	assert.NotNil(t, got)
	assert.IsType(t, trustDomainClient{}, got)
}

func TestClientCreateFederationRelationships(t *testing.T) {
	tests := []struct {
		name                    string
		input                   []*FederationRelationship
		expected                []*FederationRelationshipResult
		err                     string
		federationRelationships []*FederationRelationship
		clientResponse          *trustdomainv1.BatchCreateFederationRelationshipResponse
		clientErr               string
	}{
		{
			name:     "ok_https",
			input:    []*FederationRelationship{inValidHttpsSpiffeExample},
			expected: []*FederationRelationshipResult{outValidHttpsSpiffeExample},
			clientResponse: &trustdomainv1.BatchCreateFederationRelationshipResponse{
				Results: []*trustdomainv1.BatchCreateFederationRelationshipResponse_Result{
					{
						Status:                 &types.Status{Code: int32(codes.OK)},
						FederationRelationship: makeFederationRelationship("example.org", &HTTPSSpiffeBundleEndpointProfile{}),
					},
				},
			},
		}, {
			name:     "ok_web",
			input:    []*FederationRelationship{inValidHttpsSpiffeExample},
			expected: []*FederationRelationshipResult{outValidHttpsWebExample},
			clientResponse: &trustdomainv1.BatchCreateFederationRelationshipResponse{
				Results: []*trustdomainv1.BatchCreateFederationRelationshipResponse_Result{
					{
						Status:                 &types.Status{Code: int32(codes.OK)},
						FederationRelationship: makeFederationRelationship("example.org", &HTTPSWebBundleEndpointProfile{}),
					},
				},
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
				Results: []*trustdomainv1.BatchCreateFederationRelationshipResponse_Result{
					{
						Status:                 &types.Status{Code: int32(codes.OK)},
						FederationRelationship: &types.FederationRelationship{},
					},
				},
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

			spireTrustDomainClient.batchCreateFederationRelationshipsReponse = tt.clientResponse
			if tt.clientErr != "" {
				spireTrustDomainClient.batchCreateFederationRelationshipsError = errors.New(tt.clientErr)
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

func TestClientUpdateFederationRelationships(t *testing.T) {
	tests := []struct {
		name           string
		input          []*FederationRelationship
		expected       []*FederationRelationshipResult
		err            string
		clientResponse *trustdomainv1.BatchUpdateFederationRelationshipResponse
		clientErr      string
	}{
		{
			name:     "ok_spiffe",
			input:    []*FederationRelationship{inValidHttpsSpiffeExample},
			expected: []*FederationRelationshipResult{outValidHttpsSpiffeExample},
			clientResponse: &trustdomainv1.BatchUpdateFederationRelationshipResponse{
				Results: []*trustdomainv1.BatchUpdateFederationRelationshipResponse_Result{
					{
						Status:                 &types.Status{Code: int32(codes.OK)},
						FederationRelationship: makeFederationRelationship("example.org", &HTTPSSpiffeBundleEndpointProfile{}),
					},
				},
			},
		},
		{
			name:     "ok_web",
			input:    []*FederationRelationship{inValidHttpsSpiffeExample},
			expected: []*FederationRelationshipResult{outValidHttpsWebExample},
			clientResponse: &trustdomainv1.BatchUpdateFederationRelationshipResponse{
				Results: []*trustdomainv1.BatchUpdateFederationRelationshipResponse_Result{
					{
						Status:                 &types.Status{Code: int32(codes.OK)},
						FederationRelationship: makeFederationRelationship("example.org", &HTTPSWebBundleEndpointProfile{}),
					},
				},
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
			err:       "failed to update federation relationships: error_from_client",
		},
		{
			name:  "invalid_client_response",
			input: []*FederationRelationship{inValidHttpsSpiffeExample},
			clientResponse: &trustdomainv1.BatchUpdateFederationRelationshipResponse{
				Results: []*trustdomainv1.BatchUpdateFederationRelationshipResponse_Result{
					{
						Status:                 &types.Status{Code: int32(codes.OK)},
						FederationRelationship: &types.FederationRelationship{},
					},
				},
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

			spireTrustDomainClient.batchUpdateFederationRelationshipsReponse = tt.clientResponse
			if tt.clientErr != "" {
				spireTrustDomainClient.batchUpdateFederationRelationshipsError = errors.New(tt.clientErr)
			}

			client := &trustDomainClient{client: spireTrustDomainClient}

			got, err := client.UpdateFederationRelationships(context.Background(), tt.input)

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

func TestClientDeleteFederationRelationships(t *testing.T) {
	tests := []struct {
		name           string
		input          []string
		expected       []*FederationRelationshipResult
		err            string
		clientResponse *trustdomainv1.BatchDeleteFederationRelationshipResponse
		clientErr      string
	}{
		{
			name:     "ok",
			input:    []string{"example.org"},
			expected: []*FederationRelationshipResult{outDeleteStatusExample},
			clientResponse: &trustdomainv1.BatchDeleteFederationRelationshipResponse{
				Results: []*trustdomainv1.BatchDeleteFederationRelationshipResponse_Result{
					{
						Status:      &types.Status{Code: int32(codes.OK)},
						TrustDomain: "example.org",
					},
				},
			},
		},
		{
			name:      "client_error",
			input:     []string{"example.org"},
			clientErr: "error_from_client",
			err:       "failed to delete federation relationships: error_from_client",
		},
		{
			name:  "invalid_client_response",
			input: []string{"example.org"},
			clientResponse: &trustdomainv1.BatchDeleteFederationRelationshipResponse{
				Results: []*trustdomainv1.BatchDeleteFederationRelationshipResponse_Result{{}},
			},
			err: "failed to parse federation relationship results: invalid proto response: ",
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			spireTrustDomainClient := &fakeSpireTrustDomainClient{}
			spireTrustDomainClient.batchDeleteFederationRelationshipsReponse = tt.clientResponse

			if tt.clientErr != "" {
				spireTrustDomainClient.batchDeleteFederationRelationshipsError = errors.New(tt.clientErr)
			}

			client := &trustDomainClient{client: spireTrustDomainClient}

			tds, err := stringsToTrustDomains(tt.input)
			assert.NoError(t, err)

			got, err := client.DeleteFederationRelationships(context.Background(), tds)

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
