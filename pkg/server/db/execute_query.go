package db

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/HewlettPackard/galadriel/pkg/common/entity"
	"github.com/HewlettPackard/galadriel/pkg/server/db/criteria"
	"github.com/Masterminds/squirrel"
	"github.com/google/uuid"
)

// ExecuteListRelationshipsQuery executes a query to retrieve relationships from the database based on the provided criteria.
// The function constructs the SQL query based on the provided criteria, including pagination, filtering by consent status,
// filtering by trust domain ID, and ordering by created at. If the listCriteria parameter is nil, the function returns
// all relationships without any filtering or ordering.
func ExecuteListRelationshipsQuery(ctx context.Context, db *sql.DB, listCriteria *criteria.ListRelationshipsCriteria, dbType Engine) (*sql.Rows, error) {
	query := squirrel.Select("*").From("relationships")

	if listCriteria != nil {
		query = applyWhereClause(query, listCriteria, dbType)
		query = applyPagination(query, listCriteria)
	}

	return buildAndExecute(ctx, db, query)
}

// ExecuteListTrustDomainQuery executes a query to retrieve trust domains from the database based on the provided criteria.
// The function constructs the SQL query based on the provided criteria, including pagination,
// and ordering by created at. If the listCriteria parameter is nil, the function returns
// all trust domains without any filtering or ordering.
func ExecuteListTrustDomainQuery(ctx context.Context, db *sql.DB, listCriteria *criteria.ListTrustDomainCriteria) (*sql.Rows, error) {
	query := squirrel.Select("*").From("trust_domains")

	if listCriteria != nil {
		query = applyPagination(query, listCriteria)
	}

	return buildAndExecute(ctx, db, query)
}

func applyPagination(query squirrel.SelectBuilder, listCriteria criteria.Criteria) squirrel.SelectBuilder {
	// Ensuring uint types for operations bellow
	offset := uint(0)
	pageSize := uint(0)

	offset = (listCriteria.GetPageNumber() - 1) * listCriteria.GetPageSize()
	pageSize = listCriteria.GetPageSize()

	if listCriteria.GetOrderDirection() != criteria.NoOrder {
		query = query.OrderBy(fmt.Sprintf("created_at %s", listCriteria.GetOrderDirection()))
	}

	if pageSize > 0 {
		query = query.Limit(uint64(pageSize)).Offset(uint64(offset))
	}

	return query
}

func buildAndExecute(ctx context.Context, db *sql.DB, query squirrel.SelectBuilder) (*sql.Rows, error) {
	toSql, args, err := query.ToSql()
	if err != nil {
		return nil, fmt.Errorf("failed to build SQL query: %w", err)
	}

	rows, err := db.QueryContext(ctx, toSql, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to execute SQL query: %w", err)
	}

	return rows, nil
}

func applyWhereClause(query squirrel.SelectBuilder, listCriteria *criteria.ListRelationshipsCriteria, dbType Engine) squirrel.SelectBuilder {

	if listCriteria.FilterByConsentStatus == nil && !listCriteria.FilterByTrustDomainID.Valid {
		return query
	}

	conditions := squirrel.And{}

	if listCriteria.FilterByConsentStatus != nil && listCriteria.FilterByTrustDomainID.Valid {
		consentCondition := buildConsentConditionByTrustDomainID(*listCriteria.FilterByConsentStatus, listCriteria.FilterByTrustDomainID.UUID, dbType)
		conditions = append(conditions, consentCondition)
	} else {
		if listCriteria.FilterByConsentStatus != nil {
			consentCondition := buildConsentCondition(*listCriteria.FilterByConsentStatus, dbType)
			conditions = append(conditions, consentCondition)
		}

		if listCriteria.FilterByTrustDomainID.Valid {
			trustDomainIDCondition := buildTrustDomainIDCondition(listCriteria.FilterByTrustDomainID.UUID, dbType)
			conditions = append(conditions, trustDomainIDCondition)
		}
	}

	return query.Where(conditions)
}

// The following functions use a different syntax for Postgres due to an issue with Squirrel library:
// https://github.com/Masterminds/squirrel/issues/358
func buildConsentCondition(consentStatus entity.ConsentStatus, dbType Engine) squirrel.Sqlizer {
	if dbType == Postgres {
		return squirrel.Or{
			squirrel.Expr("trust_domain_a_consent = $1 OR trust_domain_b_consent = $2", consentStatus, consentStatus),
		}
	}
	return squirrel.Or{
		squirrel.Eq{"trust_domain_a_consent": consentStatus},
		squirrel.Eq{"trust_domain_b_consent": consentStatus},
	}
}

func buildTrustDomainIDCondition(trustDomainID uuid.UUID, dbType Engine) squirrel.Sqlizer {
	if dbType == Postgres {
		return squirrel.Or{
			squirrel.Expr("trust_domain_a_id = $1 OR trust_domain_b_id = $2", trustDomainID, trustDomainID),
		}
	}
	return squirrel.Or{
		squirrel.Eq{"trust_domain_a_id": trustDomainID},
		squirrel.Eq{"trust_domain_b_id": trustDomainID},
	}
}

func buildConsentConditionByTrustDomainID(consentStatus entity.ConsentStatus, trustDomainID uuid.UUID, dbType Engine) squirrel.Sqlizer {
	if dbType == Postgres {
		return squirrel.Expr(
			"(trust_domain_a_id = $1 AND trust_domain_a_consent = $2) OR (trust_domain_b_id = $3 AND trust_domain_b_consent = $4)",
			trustDomainID, consentStatus, trustDomainID, consentStatus,
		)
	}
	return squirrel.Or{
		squirrel.And{squirrel.Eq{"trust_domain_a_id": trustDomainID}, squirrel.Eq{"trust_domain_a_consent": consentStatus}},
		squirrel.And{squirrel.Eq{"trust_domain_b_id": trustDomainID}, squirrel.Eq{"trust_domain_b_consent": consentStatus}},
	}
}
