package branch

import (
	"context"
	"database/sql"
	"time"
)

// SQLiteRepository implements Repository for SQLite
type SQLiteRepository struct {
	db *sql.DB
}

func NewSQLiteRepository(db *sql.DB) *SQLiteRepository {
	return &SQLiteRepository{db: db}
}

const branchColumns = `id, product_id, type, name, display_name, source_branch_line_id, forked_from_version_id, status, created_at, updated_at, closed_at`

func (r *SQLiteRepository) Create(ctx context.Context, b *BranchLine) error {
	var closedAt *string
	if b.ClosedAt != nil {
		s := b.ClosedAt.Format(time.RFC3339)
		closedAt = &s
	}
	_, err := r.db.ExecContext(ctx,
		`INSERT INTO branch_line (`+branchColumns+`)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		b.ID, b.ProductID, string(b.Type), b.Name, b.DisplayName,
		b.SourceBranchLineID, b.ForkedFromVersionID, string(b.Status),
		b.CreatedAt.Format(time.RFC3339), b.UpdatedAt.Format(time.RFC3339), closedAt)
	return err
}

func (r *SQLiteRepository) FindByID(ctx context.Context, id string) (*BranchLine, error) {
	row := r.db.QueryRowContext(ctx,
		`SELECT `+branchColumns+` FROM branch_line WHERE id = ?`, id)
	return scanBranch(row)
}

func (r *SQLiteRepository) FindByProductID(ctx context.Context, productID string, opts ListOptions) ([]*BranchLine, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT `+branchColumns+` FROM branch_line WHERE product_id = ? ORDER BY created_at LIMIT ? OFFSET ?`,
		productID, opts.Limit, opts.Offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var branches []*BranchLine
	for rows.Next() {
		b, err := scanBranchRows(rows)
		if err != nil {
			return nil, err
		}
		branches = append(branches, b)
	}
	return branches, rows.Err()
}

func (r *SQLiteRepository) FindMainByProductID(ctx context.Context, productID string) (*BranchLine, error) {
	row := r.db.QueryRowContext(ctx,
		`SELECT `+branchColumns+` FROM branch_line WHERE product_id = ? AND type = 'MAIN'`, productID)
	return scanBranch(row)
}

func (r *SQLiteRepository) Update(ctx context.Context, b *BranchLine) error {
	var closedAt *string
	if b.ClosedAt != nil {
		s := b.ClosedAt.Format(time.RFC3339)
		closedAt = &s
	}
	_, err := r.db.ExecContext(ctx,
		`UPDATE branch_line SET display_name = ?, status = ?, updated_at = ?, closed_at = ? WHERE id = ?`,
		b.DisplayName, string(b.Status), b.UpdatedAt.Format(time.RFC3339), closedAt, b.ID)
	return err
}

// --- scan helpers ---

func scanBranch(row *sql.Row) (*BranchLine, error) {
	b, err := scanBranchInto(row)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return b, err
}

func scanBranchRows(rows *sql.Rows) (*BranchLine, error) {
	return scanBranchInto(rows)
}

type scanner interface {
	Scan(dest ...any) error
}

func scanBranchInto(s scanner) (*BranchLine, error) {
	var b BranchLine
	var branchType, status, createdAt, updatedAt string
	var closedAt *string
	err := s.Scan(&b.ID, &b.ProductID, &branchType, &b.Name, &b.DisplayName,
		&b.SourceBranchLineID, &b.ForkedFromVersionID,
		&status, &createdAt, &updatedAt, &closedAt)
	if err != nil {
		return nil, err
	}
	b.Type = BranchType(branchType)
	b.Status = BranchStatus(status)
	b.CreatedAt, _ = time.Parse(time.RFC3339, createdAt)
	b.UpdatedAt, _ = time.Parse(time.RFC3339, updatedAt)
	if closedAt != nil {
		t, _ := time.Parse(time.RFC3339, *closedAt)
		b.ClosedAt = &t
	}
	return &b, nil
}
