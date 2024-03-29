package db

import (
	"context"

	"github.com/HewlettPackard/galadriel/pkg/common/entity"
	"github.com/HewlettPackard/galadriel/pkg/server/db/criteria"
	"github.com/google/uuid"
	"github.com/spiffe/go-spiffe/v2/spiffeid"
)

type Datastore interface {
	CreateOrUpdateTrustDomain(ctx context.Context, req *entity.TrustDomain) (*entity.TrustDomain, error)
	DeleteTrustDomain(ctx context.Context, trustDomainID uuid.UUID) error
	FindTrustDomainByID(ctx context.Context, trustDomainID uuid.UUID) (*entity.TrustDomain, error)
	FindTrustDomainByName(ctx context.Context, trustDomain spiffeid.TrustDomain) (*entity.TrustDomain, error)
	ListTrustDomains(ctx context.Context, criteria *criteria.ListTrustDomainsCriteria) ([]*entity.TrustDomain, error)

	CreateOrUpdateBundle(ctx context.Context, req *entity.Bundle) (*entity.Bundle, error)
	DeleteBundle(ctx context.Context, bundleID uuid.UUID) error
	FindBundleByID(ctx context.Context, bundleID uuid.UUID) (*entity.Bundle, error)
	FindBundleByTrustDomainID(ctx context.Context, trustDomainID uuid.UUID) (*entity.Bundle, error)
	ListBundles(ctx context.Context) ([]*entity.Bundle, error)

	CreateJoinToken(ctx context.Context, req *entity.JoinToken) (*entity.JoinToken, error)
	DeleteJoinToken(ctx context.Context, joinTokenID uuid.UUID) error
	FindJoinToken(ctx context.Context, token string) (*entity.JoinToken, error)
	FindJoinTokensByID(ctx context.Context, joinTokenID uuid.UUID) (*entity.JoinToken, error)
	UpdateJoinToken(ctx context.Context, joinTokenID uuid.UUID, used bool) (*entity.JoinToken, error)
	FindJoinTokensByTrustDomainID(ctx context.Context, trustDomainID uuid.UUID) ([]*entity.JoinToken, error)
	ListJoinTokens(ctx context.Context) ([]*entity.JoinToken, error)

	CreateOrUpdateRelationship(ctx context.Context, req *entity.Relationship) (*entity.Relationship, error)
	DeleteRelationship(ctx context.Context, relationshipID uuid.UUID) error
	FindRelationshipByID(ctx context.Context, relationshipID uuid.UUID) (*entity.Relationship, error)
	FindRelationshipsByTrustDomainID(ctx context.Context, trustDomainID uuid.UUID) ([]*entity.Relationship, error)
	ListRelationships(ctx context.Context, criteria *criteria.ListRelationshipsCriteria) ([]*entity.Relationship, error)
}
