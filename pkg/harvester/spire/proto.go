package spire

import (
	"crypto"
	"crypto/x509"
	"fmt"
	"time"

	"github.com/spiffe/go-spiffe/v2/bundle/spiffebundle"
	"github.com/spiffe/go-spiffe/v2/spiffeid"
	apitypes "github.com/spiffe/spire-api-sdk/proto/spire/api/types"
)

func protoToBundle(in *apitypes.Bundle) (*spiffebundle.Bundle, error) {
	if in == nil {
		return nil, fmt.Errorf("bundle is empty")
	}

	td, err := spiffeid.TrustDomainFromString(in.TrustDomain)
	if err != nil {
		return nil, fmt.Errorf("failed to parse trust domain: %v", err)
	}

	x509authorities, err := protoToX509Certificates(in.X509Authorities)
	if err != nil {
		return nil, fmt.Errorf("failed to parse x509 authorities: %v", err)
	}

	jwtAuthorities, err := protoToJWTAuthorities(in.JwtAuthorities)
	if err != nil {
		return nil, fmt.Errorf("failed to parse jwt authorities: %v", err)
	}

	out := spiffebundle.New(td)

	out.SetX509Authorities(x509authorities)
	out.SetJWTAuthorities(jwtAuthorities)

	if in.RefreshHint != 0 {
		out.SetRefreshHint(time.Duration(in.RefreshHint) * time.Second)
	}
	if in.SequenceNumber != 0 {
		out.SetSequenceNumber(in.SequenceNumber)
	}

	return out, nil
}

func protoToX509Certificates(in []*apitypes.X509Certificate) ([]*x509.Certificate, error) {
	var out []*x509.Certificate

	for _, sc := range in {
		cert, err := x509.ParseCertificate(sc.Asn1)
		if err != nil {
			return nil, err
		}
		out = append(out, cert)
	}

	return out, nil
}

func protoToJWTAuthorities(in []*apitypes.JWTKey) (map[string]crypto.PublicKey, error) {
	out := make(map[string]crypto.PublicKey, len(in))

	for _, sjwt := range in {
		pub, err := x509.ParsePKIXPublicKey(sjwt.PublicKey)
		if err != nil {
			return nil, fmt.Errorf("failed to parse public key id %s: %v", sjwt.KeyId, err)
		}
		out[sjwt.KeyId] = pub
	}

	return out, nil
}
