package harvester

import (
	"fmt"

	"github.com/HewlettPackard/galadriel/pkg/common/entity"
	"github.com/spiffe/go-spiffe/v2/spiffeid"
)

func (b BundlePut) ToEntity() (*entity.Bundle, error) {
	td, err := spiffeid.TrustDomainFromString(b.TrustDomain)
	if err != nil {
		return nil, fmt.Errorf("malformed trust domain[%v]: %w", b.TrustDomain, err)
	}

	var sig []byte
	if b.Signature != nil {
		sig = []byte(*b.Signature)
	}

	var cert []byte
	if b.SigningCertificate != nil {
		cert = []byte(*b.SigningCertificate)
	}

	return &entity.Bundle{
		Data:               []byte(b.TrustBundle),
		Digest:             []byte(b.Digest),
		Signature:          sig,
		TrustDomainName:    td,
		SigningCertificate: cert,
	}, nil
}
