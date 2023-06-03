package db

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/HewlettPackard/galadriel/pkg/server/db/options"
	"github.com/Masterminds/squirrel"
)

func ExecuteRelationshipsQuery(ctx context.Context, db *sql.DB, opts *options.ListRelationshipsCriteria) (*sql.Rows, error) {
	offset := (opts.PageNumber - 1) * opts.PageSize
	query := squirrel.Select("*").From("relationships")

	if opts.FilterByConsentStatus != nil {
		query = query.Where(
			fmt.Sprintf(
				"(trust_domain_a_consent = '%s' OR trust_domain_b_consent = '%s')",
				*opts.FilterByConsentStatus,
				*opts.FilterByConsentStatus,
			),
		)
	}

	if opts.OrderByCreatedAt != options.NotSet {
		query = query.OrderBy(fmt.Sprintf("created_at %s", opts.OrderByCreatedAt))
	}

	if opts.PageSize > 0 {
		query = query.Limit(uint64(opts.PageSize)).Offset(uint64(offset))
	}

	sql, args, err := query.ToSql()
	if err != nil {
		return nil, fmt.Errorf("failed to build SQL query: %w", err)
	}

	rows, err := db.QueryContext(ctx, sql, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to execute SQL query: %w", err)
	}
	return rows, nil
}
