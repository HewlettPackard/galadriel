package spire

import (
	"crypto"
	"crypto/x509"
	"fmt"

	"github.com/spiffe/go-spiffe/v2/bundle/spiffebundle"
	"github.com/spiffe/go-spiffe/v2/spiffeid"
	apitypes "github.com/spiffe/spire-api-sdk/proto/spire/api/types"
)

func protoToBundle(in *apitypes.Bundle) (*spiffebundle.Bundle, error) {
	td, err := spiffeid.TrustDomainFromString(in.TrustDomain)
	if err != nil {
		return nil, err
	}

	x509authorities, err := protoToX509Certificates(in.X509Authorities)
	if err != nil {
		return nil, err
	}

	jwtAuthorities, err := protoToJWTAuthorities(in.JwtAuthorities)
	if err != nil {
		return nil, err
	}

	out := spiffebundle.New(td)

	out.SetX509Authorities(x509authorities)
	out.SetJWTAuthorities(jwtAuthorities)

	return out, nil
}

func protoToX509Certificates(in []*apitypes.X509Certificate) ([]*x509.Certificate, error) {
	out := make([]*x509.Certificate, len(in))

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
			return nil, err
		}
		out[sjwt.KeyId] = pub
	}

	return out, nil
}

func federationRelationshipsToProto(in []FederationRelationship) []*apitypes.FederationRelationship {
	out := make([]*apitypes.FederationRelationship, len(in))

	for _, inRel := range in {
		tdBundle, err := bundleToProto(inRel.TrustDomainBundle)
		if err != nil {
			return nil //,fmt.Errof("failed to convert trust domain bundle to proto: %v", err)
		}

		outRel := &apitypes.FederationRelationship{
			TrustDomain:       inRel.TrustDomain.String(),
			BundleEndpointUrl: inRel.BundleEndpointURL,
			TrustDomainBundle: tdBundle,
		}

		outRel.TrustDomainBundle = &apitypes.Bundle{}

		switch inRel.BundleEndpointProfile.(type) {
		case HTTPSWebBundleEndpointProfile:
			outRel.BundleEndpointProfile = &apitypes.FederationRelationship_HttpsWeb{
				HttpsWeb: &apitypes.HTTPSWebProfile{},
			}
		case HTTPSSpiffeBundleEndpointProfile:
			outRel.BundleEndpointProfile = &apitypes.FederationRelationship_HttpsSpiffe{
				HttpsSpiffe: &apitypes.HTTPSSPIFFEProfile{
					EndpointSpiffeId: inRel.BundleEndpointProfile.(HTTPSSpiffeBundleEndpointProfile).SpiffeID.String(),
				},
			}
		}

	}

	return out
}

func bundleToProto(in *spiffebundle.Bundle) (*apitypes.Bundle, error) {
	out := &apitypes.Bundle{
		TrustDomain: in.TrustDomain().String(),
	}
	x509Auths, err := x509AuthoritiesToProto(in.X509Authorities())
	if err != nil {
		return nil, err
	}
	out.X509Authorities = x509Auths

	jwtAuths, err := jwtAuthoritiesToProto(in.JWTAuthorities())
	if err != nil {
		return nil, err
	}
	out.JwtAuthorities = jwtAuths

	return out, nil
}

func x509AuthoritiesToProto(in []*x509.Certificate) ([]*apitypes.X509Certificate, error) {
	out := make([]*apitypes.X509Certificate, len(in))

	for _, c := range out {
		if c.Asn1 == nil || len(c.Asn1) == 0 {
			return nil, fmt.Errorf("x509 certificate is missing ASN1 data")
		}

		out = append(out, &apitypes.X509Certificate{Asn1: c.Asn1})
	}

	return out, nil
}

func jwtAuthoritiesToProto(in map[string]crypto.PublicKey) ([]*apitypes.JWTKey, error) {
	out := make([]*apitypes.JWTKey, len(in))

	for k, v := range in {
		if k == "" {
			return nil, fmt.Errorf("JWT key is missing kid")
		}
		pk, err := x509.MarshalPKIXPublicKey(v)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal public key: %v", err)
		}

		jwt := &apitypes.JWTKey{
			PublicKey: pk,
			KeyId:     k,
		}

		out = append(out, jwt)
	}

	return out, nil
}
