package spire

import (
	"context"
	"crypto"
	"crypto/x509"
	"fmt"

	"github.com/spiffe/go-spiffe/v2/bundle/spiffebundle"
	"github.com/spiffe/go-spiffe/v2/spiffeid"
	bundlev1 "github.com/spiffe/spire-api-sdk/proto/spire/api/server/bundle/v1"
	"github.com/spiffe/spire-api-sdk/proto/spire/api/types"
	"google.golang.org/grpc"
)

type BundleClient interface {
	GetBundle(context.Context) (*spiffebundle.Bundle, error)
}

func NewBundleClient(cc grpc.ClientConnInterface) BundleClient {
	return bundleClient{client: bundlev1.NewBundleClient(cc)}
}

type bundleClient struct {
	client bundlev1.BundleClient
}

func (c bundleClient) GetBundle(ctx context.Context) (*spiffebundle.Bundle, error) {
	bundle, err := c.client.GetBundle(ctx, &bundlev1.GetBundleRequest{})
	if err != nil {
		return nil, fmt.Errorf("failed to get bundle: %w", err)
	}

	spiffeBundle, err := parseBundle(bundle)
	if err != nil {
		return nil, fmt.Errorf("failed to parse spire server bundle response: %w", err)
	}

	return spiffeBundle, nil
}

func parseBundle(grpcBundle *types.Bundle) (*spiffebundle.Bundle, error) {
	td, err := spiffeid.TrustDomainFromString(grpcBundle.TrustDomain)
	if err != nil {
		return nil, err
	}

	x509authorities, err := parseX509Certificates(grpcBundle.X509Authorities)
	if err != nil {
		return nil, err
	}

	jwtAuthorities, err := parseJWTAuthorities(grpcBundle.JwtAuthorities)
	if err != nil {
		return nil, err
	}

	bundle := spiffebundle.New(td)

	bundle.SetX509Authorities(x509authorities)
	bundle.SetJWTAuthorities(jwtAuthorities)

	return bundle, nil
}

func parseX509Certificates(spireCerts []*types.X509Certificate) ([]*x509.Certificate, error) {
	x509certs := make([]*x509.Certificate, len(spireCerts))

	for _, sc := range spireCerts {
		cert, err := x509.ParseCertificate(sc.Asn1)
		if err != nil {
			return nil, err
		}
		x509certs = append(x509certs, cert)
	}

	return x509certs, nil
}

func parseJWTAuthorities(spireJwts []*types.JWTKey) (map[string]crypto.PublicKey, error) {
	jwtAuths := make(map[string]crypto.PublicKey, len(spireJwts))

	for _, sjwt := range spireJwts {
		pub, err := x509.ParsePKIXPublicKey(sjwt.PublicKey)
		if err != nil {
			return nil, err
		}
		jwtAuths[sjwt.KeyId] = pub
	}

	return jwtAuths, nil
}
