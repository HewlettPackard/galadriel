package harvester

import (
	"github.com/HewlettPackard/galadriel/pkg/common/entity"
	"github.com/spiffe/go-spiffe/v2/spiffeid"
)

func (b BundlePut) ToEntity() *entity.Bundle {
	return &entity.Bundle{
		Data:      []byte(b.TrustBundle),
		Signature: []byte(b.Signature),
		// TODO: do we need to store it in PEM or DER?
		SigningCertificate: []byte(b.SigningCertificate),
		TrustDomainName:    spiffeid.RequireTrustDomainFromString(b.TrustDomain),
	}
}
