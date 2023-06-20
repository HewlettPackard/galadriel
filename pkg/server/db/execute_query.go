package db

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/HewlettPackard/galadriel/pkg/server/db/criteria"
	"github.com/HewlettPackard/galadriel/pkg/server/db/dbtypes"
	"github.com/Masterminds/squirrel"
)

// ExecuteListRelationshipsQuery executes a query to retrieve relationships from the database based on the provided criteria.
func ExecuteListRelationshipsQuery(ctx context.Context, db *sql.DB, listCriteria *criteria.ListRelationshipsCriteria, dbType dbtypes.Engine) (*sql.Rows, error) {
	query := squirrel.Select("*").From("relationships")

	if listCriteria != nil {
		query = applyWhereClause(query, listCriteria, dbType)
		query = applyPaginationAndOrder(query, listCriteria)
	}

	return buildAndExecute(ctx, db, query)
}

// ExecuteListTrustDomainQuery executes a query to retrieve trust domains from the database based on the provided criteria.
func ExecuteListTrustDomainQuery(ctx context.Context, db *sql.DB, listCriteria *criteria.ListTrustDomainsCriteria, dbType dbtypes.Engine) (*sql.Rows, error) {
	query := squirrel.Select("*").From("trust_domains")

	if listCriteria != nil {
		query = applyWhereClause(query, listCriteria, dbType)
		query = applyPaginationAndOrder(query, listCriteria)
	}

	return buildAndExecute(ctx, db, query)
}

func applyPaginationAndOrder(query squirrel.SelectBuilder, listCriteria criteria.QueryCriteria) squirrel.SelectBuilder {
	// Ensuring uint types for operations below
	offset := uint(0)
	pageSize := uint(0)

	order := listCriteria.GetOrderDirection()

	pageSize = listCriteria.GetPageSize()
	offset = (listCriteria.GetPageNumber() - 1) * pageSize

	if order != criteria.NoOrder {
		query = query.OrderBy(fmt.Sprintf("created_at %s", order))
	}

	if pageSize > 0 {
		query = query.Limit(uint64(pageSize)).Offset(uint64(offset))
	}

	return query
}

func applyWhereClause(query squirrel.SelectBuilder, listCriteria criteria.QueryCriteria, dbType dbtypes.Engine) squirrel.SelectBuilder {
	filters := listCriteria.GetFilters()
	if len(filters) == 0 {
		return query
	}

	conditions := squirrel.And{}
	for _, filter := range filters {
		conditions = append(conditions, filter.GetCondition(dbType))
	}

	return query.Where(conditions)
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
