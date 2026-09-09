package graph

import (
	"context"
	"database/sql"
)

// SQLiteLineageSource reads versions and branches from RC DB for lineage/timeline.
type SQLiteLineageSource struct {
	db *sql.DB
}

func NewSQLiteLineageSource(db *sql.DB) *SQLiteLineageSource {
	return &SQLiteLineageSource{db: db}
}

func (s *SQLiteLineageSource) ProductExists(ctx context.Context, productID string) (bool, error) {
	var one int
	err := s.db.QueryRowContext(ctx, `SELECT 1 FROM product WHERE id = ?`, productID).Scan(&one)
	if err == sql.ErrNoRows {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return true, nil
}

func (s *SQLiteLineageSource) BranchExists(ctx context.Context, branchID string) (bool, error) {
	var one int
	err := s.db.QueryRowContext(ctx, `SELECT 1 FROM branch_line WHERE id = ?`, branchID).Scan(&one)
	if err == sql.ErrNoRows {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return true, nil
}

const lineageSelect = `
SELECT v.id, v.branch_line_id, b.name, b.type, v.version_string, v.status,
       v.parent_version_id, v.forked_from_version_id, v.release_date, v.created_at
FROM version v
JOIN branch_line b ON b.id = v.branch_line_id`

func (s *SQLiteLineageSource) ListVersionsByProduct(ctx context.Context, productID string) ([]VersionRow, error) {
	rows, err := s.db.QueryContext(ctx,
		lineageSelect+` WHERE b.product_id = ?`, productID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanRows(rows)
}

func (s *SQLiteLineageSource) ListVersionsByBranch(ctx context.Context, branchID string) ([]VersionRow, error) {
	rows, err := s.db.QueryContext(ctx,
		lineageSelect+` WHERE v.branch_line_id = ?`, branchID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanRows(rows)
}

func scanRows(rows *sql.Rows) ([]VersionRow, error) {
	var out []VersionRow
	for rows.Next() {
		var r VersionRow
		if err := rows.Scan(&r.VersionID, &r.BranchLineID, &r.BranchName, &r.BranchType,
			&r.VersionString, &r.Status, &r.ParentVersionID, &r.ForkedFromVersionID,
			&r.ReleaseDate, &r.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, r)
	}
	return out, rows.Err()
}
