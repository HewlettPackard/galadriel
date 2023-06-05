package db

import (
	"context"

	"github.com/HewlettPackard/galadriel/pkg/common/entity"
)

// PopulateTrustDomainNames updates the TrustDomainAName and TrustDomainBName fields of each
// Relationship in the given slice based on their TrustDomainAID and TrustDomainBID, respectively.
// It fetches the trust domain names from the provided Datastore.
func PopulateTrustDomainNames(ctx context.Context, datastore Datastore, relationships ...*entity.Relationship) ([]*entity.Relationship, error) {
	for _, r := range relationships {
		tda, err := datastore.FindTrustDomainByID(ctx, r.TrustDomainAID)
		if err != nil {
			return nil, err
		}
		r.TrustDomainAName = tda.Name

		tdb, err := datastore.FindTrustDomainByID(ctx, r.TrustDomainBID)
		if err != nil {
			return nil, err
		}
		r.TrustDomainBName = tdb.Name
	}
	return relationships, nil
}
