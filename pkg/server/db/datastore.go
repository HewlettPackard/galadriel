package db

import (
	"context"

	"github.com/HewlettPackard/galadriel/pkg/common/entity"
	"github.com/HewlettPackard/galadriel/pkg/server/db/criteria"
	"github.com/google/uuid"
	"github.com/spiffe/go-spiffe/v2/spiffeid"
)

type Engine string

const (
	Postgres Engine = "postgres"
	SQLite   Engine = "sqlite3"
)

type Datastore interface {
	// Trust Domain
	DeleteTrustDomain(ctx context.Context, trustDomainID uuid.UUID) error
	FindTrustDomainByID(ctx context.Context, trustDomainID uuid.UUID) (*entity.TrustDomain, error)
	CreateOrUpdateTrustDomain(ctx context.Context, req *entity.TrustDomain) (*entity.TrustDomain, error)
	FindTrustDomainByName(ctx context.Context, trustDomain spiffeid.TrustDomain) (*entity.TrustDomain, error)
	ListTrustDomains(ctx context.Context, criteria *criteria.ListTrustDomainCriteria) ([]*entity.TrustDomain, error)

	// Bundles
	ListBundles(ctx context.Context) ([]*entity.Bundle, error)
	DeleteBundle(ctx context.Context, bundleID uuid.UUID) error
	FindBundleByID(ctx context.Context, bundleID uuid.UUID) (*entity.Bundle, error)
	CreateOrUpdateBundle(ctx context.Context, req *entity.Bundle) (*entity.Bundle, error)
	FindBundleByTrustDomainID(ctx context.Context, trustDomainID uuid.UUID) (*entity.Bundle, error)

	// Token
	ListJoinTokens(ctx context.Context) ([]*entity.JoinToken, error)
	DeleteJoinToken(ctx context.Context, joinTokenID uuid.UUID) error
	FindJoinToken(ctx context.Context, token string) (*entity.JoinToken, error)
	CreateJoinToken(ctx context.Context, req *entity.JoinToken) (*entity.JoinToken, error)
	FindJoinTokensByID(ctx context.Context, joinTokenID uuid.UUID) (*entity.JoinToken, error)
	UpdateJoinToken(ctx context.Context, joinTokenID uuid.UUID, used bool) (*entity.JoinToken, error)
	FindJoinTokensByTrustDomainID(ctx context.Context, trustDomainID uuid.UUID) ([]*entity.JoinToken, error)

	// Relationships
	DeleteRelationship(ctx context.Context, relationshipID uuid.UUID) error
	FindRelationshipByID(ctx context.Context, relationshipID uuid.UUID) (*entity.Relationship, error)
	CreateOrUpdateRelationship(ctx context.Context, req *entity.Relationship) (*entity.Relationship, error)
	FindRelationshipsByTrustDomainID(ctx context.Context, trustDomainID uuid.UUID) ([]*entity.Relationship, error)
	ListRelationships(ctx context.Context, criteria *criteria.ListRelationshipsCriteria) ([]*entity.Relationship, error)
}
