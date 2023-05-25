package harvester

import (
	"fmt"

	"github.com/HewlettPackard/galadriel/pkg/common/entity"
	"github.com/HewlettPackard/galadriel/pkg/common/util/encoding"
	"github.com/spiffe/go-spiffe/v2/spiffeid"
)

func (b PutBundleRequest) ToEntity() (*entity.Bundle, error) {
	td, err := spiffeid.TrustDomainFromString(b.TrustDomain)
	if err != nil {
		return nil, fmt.Errorf("malformed trust domain[%v]: %w", b.TrustDomain, err)
	}

	var sig []byte
	if b.Signature != nil {
		sig, err = encoding.DecodeFromBase64(*b.Signature)
		if err != nil {
			return nil, fmt.Errorf("cannot decode signature: %w", err)
		}
	}

	var dig []byte
	if b.Digest != "" {
		dig, err = encoding.DecodeFromBase64(b.Digest)
		if err != nil {
			return nil, fmt.Errorf("cannot decode digest: %w", err)
		}
	}

	var cert []byte
	if b.SigningCertificate != nil {
		cert, err = encoding.DecodeFromBase64(*b.SigningCertificate)
		if err != nil {
			return nil, fmt.Errorf("cannot decode signing certificate: %w", err)
		}
	}

	return &entity.Bundle{
		Data:               []byte(b.TrustBundle),
		Digest:             dig,
		Signature:          sig,
		TrustDomainName:    td,
		SigningCertificate: cert,
	}, nil
}
