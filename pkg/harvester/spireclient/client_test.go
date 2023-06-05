package spireclient

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"math/big"
	"testing"
	"time"

	bundlev1 "github.com/spiffe/spire-api-sdk/proto/spire/api/server/bundle/v1"
	"github.com/spiffe/spire-api-sdk/proto/spire/api/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
)

type fakeBundleClient struct {
	bundlev1.BundleClient // Embedded interface, all methods will return not implemented error by default
	err                   error
	bundle                *types.Bundle
}

func (f *fakeBundleClient) GetBundle(ctx context.Context, in *bundlev1.GetBundleRequest, opts ...grpc.CallOption) (*types.Bundle, error) {
	if f.err != nil {
		return nil, f.err
	}
	return f.bundle, nil
}

func TestGetBundle(t *testing.T) {
	// Arrange
	publicKey, keyBytes := generateJWTKey(t)
	b := &types.Bundle{
		TrustDomain: "example.org",
		X509Authorities: []*types.X509Certificate{
			{
				Asn1: generateCertDER(t), // an ASN.1 DER encoded X.509 certificate
			},
		},
		JwtAuthorities: []*types.JWTKey{
			{
				KeyId:     "test-key-id",
				PublicKey: keyBytes,
			},
		},
		RefreshHint:    3600,
		SequenceNumber: 42,
	}
	fakeClient := &fakeBundleClient{
		// fake response
		bundle: b,
	}

	client := newLocalSpireServerWithClient(fakeClient)

	// Act
	bundle, err := client.GetBundle(context.Background())
	require.NoError(t, err)

	// Assert
	assert.Equal(t, fakeClient.bundle.TrustDomain, bundle.TrustDomain().String())
	assert.Equal(t, fakeClient.bundle.X509Authorities[0].Asn1, bundle.X509Authorities()[0].Raw)
	assert.Equal(t, publicKey, bundle.JWTAuthorities()["test-key-id"])
}

// TODO add tests for other methods

func generateCertDER(t *testing.T) []byte {
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)

	// Create a self-signed certificate template
	certTemplate := &x509.Certificate{
		SerialNumber:          big.NewInt(1),
		Subject:               pkix.Name{CommonName: "example.org"},
		NotBefore:             time.Now(),
		NotAfter:              time.Now().AddDate(1, 0, 0),
		KeyUsage:              x509.KeyUsageDigitalSignature | x509.KeyUsageKeyEncipherment,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
		IsCA:                  true,
	}

	// Create a self-signed X.509 certificate using the private key and template
	derBytes, err := x509.CreateCertificate(rand.Reader, certTemplate, certTemplate, &privateKey.PublicKey, privateKey)
	require.NoError(t, err)

	return derBytes
}

func generateJWTKey(t *testing.T) (*rsa.PublicKey, []byte) {
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)

	// Extract the public key from the private key
	publicKey := &privateKey.PublicKey

	// Encode the public key in PKIX format
	pkixBytes, err := x509.MarshalPKIXPublicKey(publicKey)
	require.NoError(t, err)

	return publicKey, pkixBytes
}
