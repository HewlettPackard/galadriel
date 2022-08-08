package spire

import (
	"crypto"
	"crypto/x509"
	"fmt"

	"github.com/spiffe/go-spiffe/v2/bundle/spiffebundle"
	"github.com/spiffe/go-spiffe/v2/spiffeid"
	trustdomainv1 "github.com/spiffe/spire-api-sdk/proto/spire/api/server/trustdomain/v1"
	apitypes "github.com/spiffe/spire-api-sdk/proto/spire/api/types"
)

func protoToBundle(in *apitypes.Bundle) (*spiffebundle.Bundle, error) {
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
			return nil, fmt.Errorf("failed to parse public key id %s: %v", sjwt.KeyId, err)
		}
		out[sjwt.KeyId] = pub
	}

	return out, nil
}

func federationRelationshipsToProto(in []*FederationRelationship) ([]*apitypes.FederationRelationship, error) {
	out := make([]*apitypes.FederationRelationship, len(in))

	for _, inRel := range in {
		tdBundle, err := bundleToProto(inRel.TrustDomainBundle)
		if err != nil {
			return nil, fmt.Errorf("failed to convert trust domain bundle to proto: %v", err)
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
		default:
			return nil, fmt.Errorf("unsupported bundle endpoint profile for trust domain %s: %T", tdBundle.GetTrustDomain(), inRel.BundleEndpointProfile)
		}

	}

	return out, nil
}

func bundleToProto(in *spiffebundle.Bundle) (*apitypes.Bundle, error) {
	out := &apitypes.Bundle{
		TrustDomain: in.TrustDomain().String(),
	}
	x509Auths, err := x509AuthoritiesToProto(in.X509Authorities())
	if err != nil {
		return nil, fmt.Errorf("failed to convert x509 authorities to proto: %v", err)
	}
	out.X509Authorities = x509Auths

	jwtAuths, err := jwtAuthoritiesToProto(in.JWTAuthorities())
	if err != nil {
		return nil, fmt.Errorf("failed to convert jwt authorities to proto: %v", err)
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

// TODO: the next two functions (create and update) are the same but take a different argument.
// The third function (delete) has minor differences. Refactor to reuse code, probably with interfaces or generics.
func protoCreateToFederationRelatioshipResult(in *trustdomainv1.BatchCreateFederationRelationshipResponse) ([]*FederationRelationshipResult, error) {
	out := make([]*FederationRelationshipResult, len(in.GetResults()))

	for _, r := range in.GetResults() {
		frel, err := protoToFederationsRelationship(r.GetFederationRelationship())
		if err != nil {
			return nil, fmt.Errorf("failed to convert federation relationship to proto: %v", err)
		}
		rOut := &FederationRelationshipResult{
			status: &FederationRelationshipResultStatus{
				code:    r.GetStatus().Code,
				message: r.GetStatus().Message,
			},
			federationRelationship: frel,
		}
		out = append(out, rOut)
	}

	return out, nil
}

func protoUpdateToFederationRelatioshipResult(in *trustdomainv1.BatchUpdateFederationRelationshipResponse) ([]*FederationRelationshipResult, error) {
	out := make([]*FederationRelationshipResult, len(in.GetResults()))

	for _, r := range in.GetResults() {
		frel, err := protoToFederationsRelationship(r.GetFederationRelationship())
		if err != nil {
			return nil, fmt.Errorf("failed to convert federation relationship to proto: %v", err)
		}
		rOut := &FederationRelationshipResult{
			status: &FederationRelationshipResultStatus{
				code:    r.GetStatus().Code,
				message: r.GetStatus().Message,
			},
			federationRelationship: frel,
		}
		out = append(out, rOut)
	}

	return out, nil
}

func protoDeleteToFederationRelatioshipResult(in *trustdomainv1.BatchDeleteFederationRelationshipResponse) ([]*FederationRelationshipResult, error) {
	out := make([]*FederationRelationshipResult, len(in.GetResults()))

	for _, r := range in.GetResults() {
		rOut := &FederationRelationshipResult{
			status: &FederationRelationshipResultStatus{
				code:        r.GetStatus().Code,
				message:     r.GetStatus().Message,
				trustDomain: r.GetTrustDomain(),
			},
		}
		out = append(out, rOut)
	}

	return out, nil
}

func protoToFederationsRelationships(in *trustdomainv1.ListFederationRelationshipsResponse) ([]*FederationRelationship, error) {
	var out []*FederationRelationship
	for _, inRel := range in.FederationRelationships {
		outRel, err := protoToFederationsRelationship(inRel)
		if err != nil {
			return nil, fmt.Errorf("failed to parse federated relationship: %v", err)
		}
		out = append(out, outRel)
	}

	return out, nil
}

func protoToFederationsRelationship(in *apitypes.FederationRelationship) (*FederationRelationship, error) {
	td, err := spiffeid.TrustDomainFromString(in.TrustDomain)
	if err != nil {
		return nil, fmt.Errorf("failed to parse federated trust domain: %v", err)
	}
	bundle, err := protoToBundle(in.TrustDomainBundle)
	if err != nil {
		return nil, fmt.Errorf("failed to parse federated trust bundle: %v", err)
	}
	profile, err := protoToBundleProfile(in)
	if err != nil {
		return nil, fmt.Errorf("failed to parse federated profile: %v", err)
	}

	out := &FederationRelationship{
		TrustDomain:           td,
		TrustDomainBundle:     bundle,
		BundleEndpointURL:     in.BundleEndpointUrl,
		BundleEndpointProfile: profile,
	}

	return out, nil
}

func protoToBundleProfile(in *apitypes.FederationRelationship) (BundleEndpointProfile, error) {
	var out BundleEndpointProfile
	switch in.BundleEndpointProfile.(type) {
	case *apitypes.FederationRelationship_HttpsWeb:
		out = HTTPSWebBundleEndpointProfile{}
	case *apitypes.FederationRelationship_HttpsSpiffe:
		spiffeId, err := spiffeid.FromString(in.GetHttpsSpiffe().EndpointSpiffeId)
		if err != nil {
			return nil, err
		}
		out = HTTPSSpiffeBundleEndpointProfile{
			SpiffeID: spiffeId,
		}
	}

	return out, nil
}

func trustDomainsToStrings(in []*spiffeid.TrustDomain) ([]string, error) {
	var out []string
	for _, td := range in {
		if td == nil || td.IsZero() {
			return nil, fmt.Errorf("invalid trust domain: %v", td)
		}
		out = append(out, td.String())
	}

	return out, nil
}
