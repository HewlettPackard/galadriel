package endpoints

import (
	"github.com/HewlettPackard/galadriel/pkg/common/entity"
	"github.com/spiffe/go-spiffe/v2/spiffeid"
)

var (
	testTrustDomain = &entity.TrustDomain{
		Name:        spiffeid.RequireTrustDomainFromString("example.org"),
		Description: "Fake trust domain",
	}
)
