package spireclient

import (
	"crypto"
	"crypto/x509"
	"errors"
	"fmt"
	"time"

	"github.com/spiffe/go-spiffe/v2/bundle/spiffebundle"
	"github.com/spiffe/go-spiffe/v2/spiffeid"
	bundlev1 "github.com/spiffe/spire-api-sdk/proto/spire/api/server/bundle/v1"
	apitypes "github.com/spiffe/spire-api-sdk/proto/spire/api/types"
	"google.golang.org/grpc/codes"
)

func protoToBundle(in *apitypes.Bundle) (*spiffebundle.Bundle, error) {
	if in == nil {
		return nil, errors.New("bundle is empty")
	}

	td, err := spiffeid.TrustDomainFromString(in.TrustDomain)
	if err != nil {
		return nil, fmt.Errorf("failed to parse trust domain: %v", err)
	}

	x509authorities, err := protoToX509Certificates(in.X509Authorities)
	if err != nil {
		return nil, fmt.Errorf("failed to parse X.509 authorities: %v", err)
	}

	jwtAuthorities, err := protoToJWTAuthorities(in.JwtAuthorities)
	if err != nil {
		return nil, fmt.Errorf("failed to parse JWT authorities: %v", err)
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

func bundleToProto(in *spiffebundle.Bundle) (*apitypes.Bundle, error) {
	if in == nil {
		return nil, errors.New("trust domain bundle must be set")
	}

	if in.TrustDomain().IsZero() {
		return nil, errors.New("trust domain must be set")
	}

	x509Auths, err := x509AuthoritiesToProto(in.X509Authorities())
	if err != nil {
		return nil, fmt.Errorf("failed to convert X.509 authorities to proto: %v", err)
	}

	jwtAuths, err := jwtAuthoritiesToProto(in.JWTAuthorities())
	if err != nil {
		return nil, fmt.Errorf("failed to convert JWT authorities to proto: %v", err)
	}

	trustDomain := in.TrustDomain().String()
	sequenceNumber, _ := in.SequenceNumber()
	refreshHint, _ := in.RefreshHint()

	out := &apitypes.Bundle{
		TrustDomain:     trustDomain,
		SequenceNumber:  sequenceNumber,
		RefreshHint:     int64(refreshHint.Seconds()),
		X509Authorities: x509Auths,
		JwtAuthorities:  jwtAuths,
	}

	return out, nil
}

func bundlesToProto(bundles []*spiffebundle.Bundle) ([]*apitypes.Bundle, error) {
	var out []*apitypes.Bundle

	for _, b := range bundles {
		bundle, err := bundleToProto(b)
		if err != nil {
			return nil, err
		}
		out = append(out, bundle)
	}

	return out, nil
}

func x509AuthoritiesToProto(in []*x509.Certificate) ([]*apitypes.X509Certificate, error) {
	var out []*apitypes.X509Certificate

	for _, c := range in {
		if len(c.Raw) == 0 {
			return nil, errors.New("an X.509 certificate is missing raw data")
		}

		out = append(out, &apitypes.X509Certificate{Asn1: c.Raw})
	}

	return out, nil
}

func jwtAuthoritiesToProto(in map[string]crypto.PublicKey) ([]*apitypes.JWTKey, error) {
	var out []*apitypes.JWTKey

	for k, v := range in {
		if k == "" {
			return nil, errors.New("a JWT key is missing key id")
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

func protoToBatchSetFederatedBundleResult(in *bundlev1.BatchSetFederatedBundleResponse) ([]*BatchSetFederatedBundleStatus, error) {
	var out []*BatchSetFederatedBundleStatus

	for _, r := range in.GetResults() {
		var bundle *spiffebundle.Bundle

		if r.Bundle != nil {
			var err error
			bundle, err = protoToBundle(r.Bundle)
			if err != nil {
				return nil, err
			}
		}

		if r.Status == nil {
			return nil, errors.New("call returned no status")
		}

		bs := &BatchSetFederatedBundleStatus{
			Bundle: bundle,
			Status: &Status{
				Message: r.Status.GetMessage(),
				Code:    codes.Code(r.Status.GetCode()),
			},
		}
		out = append(out, bs)
	}

	return out, nil
}

func protoToBatchDeleteFederatedBundleResult(in *bundlev1.BatchDeleteFederatedBundleResponse) ([]*BatchDeleteFederatedBundleStatus, error) {
	var out []*BatchDeleteFederatedBundleStatus

	for _, r := range in.GetResults() {
		if r.Status == nil {
			return nil, errors.New("call returned no status")
		}

		bs := &BatchDeleteFederatedBundleStatus{
			TrustDomain: r.TrustDomain,
			Status: &Status{
				Message: r.Status.GetMessage(),
				Code:    codes.Code(r.Status.GetCode()),
			},
		}
		out = append(out, bs)
	}

	return out, nil
}

func protoToSpiffeBundles(resp *bundlev1.ListFederatedBundlesResponse) ([]*spiffebundle.Bundle, error) {
	var out []*spiffebundle.Bundle

	for _, b := range resp.Bundles {
		bundle, err := protoToBundle(b)
		if err != nil {
			return nil, err
		}

		out = append(out, bundle)
	}

	return out, nil
}
