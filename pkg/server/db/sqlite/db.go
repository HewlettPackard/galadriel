// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.17.0

package sqlite

import (
	"context"
	"database/sql"
	"fmt"
)

type DBTX interface {
	ExecContext(context.Context, string, ...interface{}) (sql.Result, error)
	PrepareContext(context.Context, string) (*sql.Stmt, error)
	QueryContext(context.Context, string, ...interface{}) (*sql.Rows, error)
	QueryRowContext(context.Context, string, ...interface{}) *sql.Row
}

func New(db DBTX) *Queries {
	return &Queries{db: db}
}

func Prepare(ctx context.Context, db DBTX) (*Queries, error) {
	q := Queries{db: db}
	var err error
	if q.createBundleStmt, err = db.PrepareContext(ctx, createBundle); err != nil {
		return nil, fmt.Errorf("error preparing query CreateBundle: %w", err)
	}
	if q.createJoinTokenStmt, err = db.PrepareContext(ctx, createJoinToken); err != nil {
		return nil, fmt.Errorf("error preparing query CreateJoinToken: %w", err)
	}
	if q.createRelationshipStmt, err = db.PrepareContext(ctx, createRelationship); err != nil {
		return nil, fmt.Errorf("error preparing query CreateRelationship: %w", err)
	}
	if q.createTrustDomainStmt, err = db.PrepareContext(ctx, createTrustDomain); err != nil {
		return nil, fmt.Errorf("error preparing query CreateTrustDomain: %w", err)
	}
	if q.deleteBundleStmt, err = db.PrepareContext(ctx, deleteBundle); err != nil {
		return nil, fmt.Errorf("error preparing query DeleteBundle: %w", err)
	}
	if q.deleteJoinTokenStmt, err = db.PrepareContext(ctx, deleteJoinToken); err != nil {
		return nil, fmt.Errorf("error preparing query DeleteJoinToken: %w", err)
	}
	if q.deleteRelationshipStmt, err = db.PrepareContext(ctx, deleteRelationship); err != nil {
		return nil, fmt.Errorf("error preparing query DeleteRelationship: %w", err)
	}
	if q.deleteTrustDomainStmt, err = db.PrepareContext(ctx, deleteTrustDomain); err != nil {
		return nil, fmt.Errorf("error preparing query DeleteTrustDomain: %w", err)
	}
	if q.findBundleByIDStmt, err = db.PrepareContext(ctx, findBundleByID); err != nil {
		return nil, fmt.Errorf("error preparing query FindBundleByID: %w", err)
	}
	if q.findBundleByTrustDomainIDStmt, err = db.PrepareContext(ctx, findBundleByTrustDomainID); err != nil {
		return nil, fmt.Errorf("error preparing query FindBundleByTrustDomainID: %w", err)
	}
	if q.findJoinTokenStmt, err = db.PrepareContext(ctx, findJoinToken); err != nil {
		return nil, fmt.Errorf("error preparing query FindJoinToken: %w", err)
	}
	if q.findJoinTokenByIDStmt, err = db.PrepareContext(ctx, findJoinTokenByID); err != nil {
		return nil, fmt.Errorf("error preparing query FindJoinTokenByID: %w", err)
	}
	if q.findJoinTokensByTrustDomainIDStmt, err = db.PrepareContext(ctx, findJoinTokensByTrustDomainID); err != nil {
		return nil, fmt.Errorf("error preparing query FindJoinTokensByTrustDomainID: %w", err)
	}
	if q.findRelationshipByIDStmt, err = db.PrepareContext(ctx, findRelationshipByID); err != nil {
		return nil, fmt.Errorf("error preparing query FindRelationshipByID: %w", err)
	}
	if q.findRelationshipsByTrustDomainIDStmt, err = db.PrepareContext(ctx, findRelationshipsByTrustDomainID); err != nil {
		return nil, fmt.Errorf("error preparing query FindRelationshipsByTrustDomainID: %w", err)
	}
	if q.findTrustDomainByIDStmt, err = db.PrepareContext(ctx, findTrustDomainByID); err != nil {
		return nil, fmt.Errorf("error preparing query FindTrustDomainByID: %w", err)
	}
	if q.findTrustDomainByNameStmt, err = db.PrepareContext(ctx, findTrustDomainByName); err != nil {
		return nil, fmt.Errorf("error preparing query FindTrustDomainByName: %w", err)
	}
	if q.listBundlesStmt, err = db.PrepareContext(ctx, listBundles); err != nil {
		return nil, fmt.Errorf("error preparing query ListBundles: %w", err)
	}
	if q.listJoinTokensStmt, err = db.PrepareContext(ctx, listJoinTokens); err != nil {
		return nil, fmt.Errorf("error preparing query ListJoinTokens: %w", err)
	}
	if q.listRelationshipsStmt, err = db.PrepareContext(ctx, listRelationships); err != nil {
		return nil, fmt.Errorf("error preparing query ListRelationships: %w", err)
	}
	if q.listTrustDomainsStmt, err = db.PrepareContext(ctx, listTrustDomains); err != nil {
		return nil, fmt.Errorf("error preparing query ListTrustDomains: %w", err)
	}
	if q.updateBundleStmt, err = db.PrepareContext(ctx, updateBundle); err != nil {
		return nil, fmt.Errorf("error preparing query UpdateBundle: %w", err)
	}
	if q.updateJoinTokenStmt, err = db.PrepareContext(ctx, updateJoinToken); err != nil {
		return nil, fmt.Errorf("error preparing query UpdateJoinToken: %w", err)
	}
	if q.updateRelationshipStmt, err = db.PrepareContext(ctx, updateRelationship); err != nil {
		return nil, fmt.Errorf("error preparing query UpdateRelationship: %w", err)
	}
	if q.updateTrustDomainStmt, err = db.PrepareContext(ctx, updateTrustDomain); err != nil {
		return nil, fmt.Errorf("error preparing query UpdateTrustDomain: %w", err)
	}
	return &q, nil
}

func (q *Queries) Close() error {
	var err error
	if q.createBundleStmt != nil {
		if cerr := q.createBundleStmt.Close(); cerr != nil {
			err = fmt.Errorf("error closing createBundleStmt: %w", cerr)
		}
	}
	if q.createJoinTokenStmt != nil {
		if cerr := q.createJoinTokenStmt.Close(); cerr != nil {
			err = fmt.Errorf("error closing createJoinTokenStmt: %w", cerr)
		}
	}
	if q.createRelationshipStmt != nil {
		if cerr := q.createRelationshipStmt.Close(); cerr != nil {
			err = fmt.Errorf("error closing createRelationshipStmt: %w", cerr)
		}
	}
	if q.createTrustDomainStmt != nil {
		if cerr := q.createTrustDomainStmt.Close(); cerr != nil {
			err = fmt.Errorf("error closing createTrustDomainStmt: %w", cerr)
		}
	}
	if q.deleteBundleStmt != nil {
		if cerr := q.deleteBundleStmt.Close(); cerr != nil {
			err = fmt.Errorf("error closing deleteBundleStmt: %w", cerr)
		}
	}
	if q.deleteJoinTokenStmt != nil {
		if cerr := q.deleteJoinTokenStmt.Close(); cerr != nil {
			err = fmt.Errorf("error closing deleteJoinTokenStmt: %w", cerr)
		}
	}
	if q.deleteRelationshipStmt != nil {
		if cerr := q.deleteRelationshipStmt.Close(); cerr != nil {
			err = fmt.Errorf("error closing deleteRelationshipStmt: %w", cerr)
		}
	}
	if q.deleteTrustDomainStmt != nil {
		if cerr := q.deleteTrustDomainStmt.Close(); cerr != nil {
			err = fmt.Errorf("error closing deleteTrustDomainStmt: %w", cerr)
		}
	}
	if q.findBundleByIDStmt != nil {
		if cerr := q.findBundleByIDStmt.Close(); cerr != nil {
			err = fmt.Errorf("error closing findBundleByIDStmt: %w", cerr)
		}
	}
	if q.findBundleByTrustDomainIDStmt != nil {
		if cerr := q.findBundleByTrustDomainIDStmt.Close(); cerr != nil {
			err = fmt.Errorf("error closing findBundleByTrustDomainIDStmt: %w", cerr)
		}
	}
	if q.findJoinTokenStmt != nil {
		if cerr := q.findJoinTokenStmt.Close(); cerr != nil {
			err = fmt.Errorf("error closing findJoinTokenStmt: %w", cerr)
		}
	}
	if q.findJoinTokenByIDStmt != nil {
		if cerr := q.findJoinTokenByIDStmt.Close(); cerr != nil {
			err = fmt.Errorf("error closing findJoinTokenByIDStmt: %w", cerr)
		}
	}
	if q.findJoinTokensByTrustDomainIDStmt != nil {
		if cerr := q.findJoinTokensByTrustDomainIDStmt.Close(); cerr != nil {
			err = fmt.Errorf("error closing findJoinTokensByTrustDomainIDStmt: %w", cerr)
		}
	}
	if q.findRelationshipByIDStmt != nil {
		if cerr := q.findRelationshipByIDStmt.Close(); cerr != nil {
			err = fmt.Errorf("error closing findRelationshipByIDStmt: %w", cerr)
		}
	}
	if q.findRelationshipsByTrustDomainIDStmt != nil {
		if cerr := q.findRelationshipsByTrustDomainIDStmt.Close(); cerr != nil {
			err = fmt.Errorf("error closing findRelationshipsByTrustDomainIDStmt: %w", cerr)
		}
	}
	if q.findTrustDomainByIDStmt != nil {
		if cerr := q.findTrustDomainByIDStmt.Close(); cerr != nil {
			err = fmt.Errorf("error closing findTrustDomainByIDStmt: %w", cerr)
		}
	}
	if q.findTrustDomainByNameStmt != nil {
		if cerr := q.findTrustDomainByNameStmt.Close(); cerr != nil {
			err = fmt.Errorf("error closing findTrustDomainByNameStmt: %w", cerr)
		}
	}
	if q.listBundlesStmt != nil {
		if cerr := q.listBundlesStmt.Close(); cerr != nil {
			err = fmt.Errorf("error closing listBundlesStmt: %w", cerr)
		}
	}
	if q.listJoinTokensStmt != nil {
		if cerr := q.listJoinTokensStmt.Close(); cerr != nil {
			err = fmt.Errorf("error closing listJoinTokensStmt: %w", cerr)
		}
	}
	if q.listRelationshipsStmt != nil {
		if cerr := q.listRelationshipsStmt.Close(); cerr != nil {
			err = fmt.Errorf("error closing listRelationshipsStmt: %w", cerr)
		}
	}
	if q.listTrustDomainsStmt != nil {
		if cerr := q.listTrustDomainsStmt.Close(); cerr != nil {
			err = fmt.Errorf("error closing listTrustDomainsStmt: %w", cerr)
		}
	}
	if q.updateBundleStmt != nil {
		if cerr := q.updateBundleStmt.Close(); cerr != nil {
			err = fmt.Errorf("error closing updateBundleStmt: %w", cerr)
		}
	}
	if q.updateJoinTokenStmt != nil {
		if cerr := q.updateJoinTokenStmt.Close(); cerr != nil {
			err = fmt.Errorf("error closing updateJoinTokenStmt: %w", cerr)
		}
	}
	if q.updateRelationshipStmt != nil {
		if cerr := q.updateRelationshipStmt.Close(); cerr != nil {
			err = fmt.Errorf("error closing updateRelationshipStmt: %w", cerr)
		}
	}
	if q.updateTrustDomainStmt != nil {
		if cerr := q.updateTrustDomainStmt.Close(); cerr != nil {
			err = fmt.Errorf("error closing updateTrustDomainStmt: %w", cerr)
		}
	}
	return err
}

func (q *Queries) exec(ctx context.Context, stmt *sql.Stmt, query string, args ...interface{}) (sql.Result, error) {
	switch {
	case stmt != nil && q.tx != nil:
		return q.tx.StmtContext(ctx, stmt).ExecContext(ctx, args...)
	case stmt != nil:
		return stmt.ExecContext(ctx, args...)
	default:
		return q.db.ExecContext(ctx, query, args...)
	}
}

func (q *Queries) query(ctx context.Context, stmt *sql.Stmt, query string, args ...interface{}) (*sql.Rows, error) {
	switch {
	case stmt != nil && q.tx != nil:
		return q.tx.StmtContext(ctx, stmt).QueryContext(ctx, args...)
	case stmt != nil:
		return stmt.QueryContext(ctx, args...)
	default:
		return q.db.QueryContext(ctx, query, args...)
	}
}

func (q *Queries) queryRow(ctx context.Context, stmt *sql.Stmt, query string, args ...interface{}) *sql.Row {
	switch {
	case stmt != nil && q.tx != nil:
		return q.tx.StmtContext(ctx, stmt).QueryRowContext(ctx, args...)
	case stmt != nil:
		return stmt.QueryRowContext(ctx, args...)
	default:
		return q.db.QueryRowContext(ctx, query, args...)
	}
}

type Queries struct {
	db                                   DBTX
	tx                                   *sql.Tx
	createBundleStmt                     *sql.Stmt
	createJoinTokenStmt                  *sql.Stmt
	createRelationshipStmt               *sql.Stmt
	createTrustDomainStmt                *sql.Stmt
	deleteBundleStmt                     *sql.Stmt
	deleteJoinTokenStmt                  *sql.Stmt
	deleteRelationshipStmt               *sql.Stmt
	deleteTrustDomainStmt                *sql.Stmt
	findBundleByIDStmt                   *sql.Stmt
	findBundleByTrustDomainIDStmt        *sql.Stmt
	findJoinTokenStmt                    *sql.Stmt
	findJoinTokenByIDStmt                *sql.Stmt
	findJoinTokensByTrustDomainIDStmt    *sql.Stmt
	findRelationshipByIDStmt             *sql.Stmt
	findRelationshipsByTrustDomainIDStmt *sql.Stmt
	findTrustDomainByIDStmt              *sql.Stmt
	findTrustDomainByNameStmt            *sql.Stmt
	listBundlesStmt                      *sql.Stmt
	listJoinTokensStmt                   *sql.Stmt
	listRelationshipsStmt                *sql.Stmt
	listTrustDomainsStmt                 *sql.Stmt
	updateBundleStmt                     *sql.Stmt
	updateJoinTokenStmt                  *sql.Stmt
	updateRelationshipStmt               *sql.Stmt
	updateTrustDomainStmt                *sql.Stmt
}

func (q *Queries) WithTx(tx *sql.Tx) *Queries {
	return &Queries{
		db:                                   tx,
		tx:                                   tx,
		createBundleStmt:                     q.createBundleStmt,
		createJoinTokenStmt:                  q.createJoinTokenStmt,
		createRelationshipStmt:               q.createRelationshipStmt,
		createTrustDomainStmt:                q.createTrustDomainStmt,
		deleteBundleStmt:                     q.deleteBundleStmt,
		deleteJoinTokenStmt:                  q.deleteJoinTokenStmt,
		deleteRelationshipStmt:               q.deleteRelationshipStmt,
		deleteTrustDomainStmt:                q.deleteTrustDomainStmt,
		findBundleByIDStmt:                   q.findBundleByIDStmt,
		findBundleByTrustDomainIDStmt:        q.findBundleByTrustDomainIDStmt,
		findJoinTokenStmt:                    q.findJoinTokenStmt,
		findJoinTokenByIDStmt:                q.findJoinTokenByIDStmt,
		findJoinTokensByTrustDomainIDStmt:    q.findJoinTokensByTrustDomainIDStmt,
		findRelationshipByIDStmt:             q.findRelationshipByIDStmt,
		findRelationshipsByTrustDomainIDStmt: q.findRelationshipsByTrustDomainIDStmt,
		findTrustDomainByIDStmt:              q.findTrustDomainByIDStmt,
		findTrustDomainByNameStmt:            q.findTrustDomainByNameStmt,
		listBundlesStmt:                      q.listBundlesStmt,
		listJoinTokensStmt:                   q.listJoinTokensStmt,
		listRelationshipsStmt:                q.listRelationshipsStmt,
		listTrustDomainsStmt:                 q.listTrustDomainsStmt,
		updateBundleStmt:                     q.updateBundleStmt,
		updateJoinTokenStmt:                  q.updateJoinTokenStmt,
		updateRelationshipStmt:               q.updateRelationshipStmt,
		updateTrustDomainStmt:                q.updateTrustDomainStmt,
	}
}
