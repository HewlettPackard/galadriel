package harvester

import (
	"github.com/HewlettPackard/galadriel/pkg/common"
	"github.com/HewlettPackard/galadriel/pkg/common/entity"
	"github.com/spiffe/go-spiffe/v2/spiffeid"
)

func (b BundlePut) ToEntity() (*entity.Bundle, error) {
	td, err := spiffeid.TrustDomainFromString(b.TrustDomain)
	if err != nil {
		return nil, common.ErrWrongTrustDomain{Cause: err}
	}

	return &entity.Bundle{
		Data:            []byte(b.TrustBundle),
		Signature:       []byte(b.Signature),
		TrustDomainName: td,
		// TODO: do we need to store it in PEM or DER?
		SigningCertificate: []byte(b.SigningCertificate),
	}, nil
}
