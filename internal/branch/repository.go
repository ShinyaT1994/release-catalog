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

func (r *SQLiteRepository) Create(ctx context.Context, b *BranchLine) error {
	var closedAt *string
	if b.ClosedAt != nil {
		s := b.ClosedAt.Format(time.RFC3339)
		closedAt = &s
	}
	_, err := r.db.ExecContext(ctx,
		`INSERT INTO branch_line (id, product_id, type, name, display_name, source_branch_line_id, forked_from_snapshot_id, status, created_at, updated_at, closed_at)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		b.ID, b.ProductID, string(b.Type), b.Name, b.DisplayName,
		b.SourceBranchLineID, b.ForkedFromSnapshotID, string(b.Status),
		b.CreatedAt.Format(time.RFC3339), b.UpdatedAt.Format(time.RFC3339), closedAt)
	return err
}

func (r *SQLiteRepository) FindByID(ctx context.Context, id string) (*BranchLine, error) {
	row := r.db.QueryRowContext(ctx,
		`SELECT id, product_id, type, name, display_name, source_branch_line_id, forked_from_snapshot_id, status, created_at, updated_at, closed_at
		 FROM branch_line WHERE id = ?`, id)
	return scanBranch(row)
}

func (r *SQLiteRepository) FindByProductID(ctx context.Context, productID string, opts ListOptions) ([]*BranchLine, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT id, product_id, type, name, display_name, source_branch_line_id, forked_from_snapshot_id, status, created_at, updated_at, closed_at
		 FROM branch_line WHERE product_id = ? ORDER BY created_at LIMIT ? OFFSET ?`,
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
		`SELECT id, product_id, type, name, display_name, source_branch_line_id, forked_from_snapshot_id, status, created_at, updated_at, closed_at
		 FROM branch_line WHERE product_id = ? AND type = 'MAIN'`, productID)
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

// --- CurrentState SQLite Repository ---

type SQLiteCurrentStateRepository struct {
	db *sql.DB
}

func NewSQLiteCurrentStateRepository(db *sql.DB) *SQLiteCurrentStateRepository {
	return &SQLiteCurrentStateRepository{db: db}
}

func (r *SQLiteCurrentStateRepository) Upsert(ctx context.Context, cs *CurrentState) error {
	_, err := r.db.ExecContext(ctx,
		`INSERT INTO branch_current_state (branch_line_id, root_dt_project_uuid, root_bom_serial_number, root_bom_version, root_bom_sha256, source_revision, updated_at)
		 VALUES (?, ?, ?, ?, ?, ?, ?)
		 ON CONFLICT(branch_line_id) DO UPDATE SET
		   root_dt_project_uuid = excluded.root_dt_project_uuid,
		   root_bom_serial_number = excluded.root_bom_serial_number,
		   root_bom_version = excluded.root_bom_version,
		   root_bom_sha256 = excluded.root_bom_sha256,
		   source_revision = excluded.source_revision,
		   updated_at = excluded.updated_at`,
		cs.BranchLineID, cs.RootDTProjectUUID, cs.RootBOMSerial,
		cs.RootBOMVersion, cs.RootBOMSHA256, cs.SourceRevision,
		cs.UpdatedAt.Format(time.RFC3339))
	return err
}

func (r *SQLiteCurrentStateRepository) FindByBranchID(ctx context.Context, branchID string) (*CurrentState, error) {
	row := r.db.QueryRowContext(ctx,
		`SELECT branch_line_id, root_dt_project_uuid, root_bom_serial_number, root_bom_version, root_bom_sha256, source_revision, updated_at
		 FROM branch_current_state WHERE branch_line_id = ?`, branchID)
	var cs CurrentState
	var updatedAt string
	err := row.Scan(&cs.BranchLineID, &cs.RootDTProjectUUID, &cs.RootBOMSerial,
		&cs.RootBOMVersion, &cs.RootBOMSHA256, &cs.SourceRevision, &updatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	cs.UpdatedAt, _ = time.Parse(time.RFC3339, updatedAt)
	return &cs, nil
}

// --- scan helpers ---

func scanBranch(row *sql.Row) (*BranchLine, error) {
	var b BranchLine
	var branchType, status, createdAt, updatedAt string
	var closedAt *string
	err := row.Scan(&b.ID, &b.ProductID, &branchType, &b.Name, &b.DisplayName,
		&b.SourceBranchLineID, &b.ForkedFromSnapshotID,
		&status, &createdAt, &updatedAt, &closedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
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

func scanBranchRows(rows *sql.Rows) (*BranchLine, error) {
	var b BranchLine
	var branchType, status, createdAt, updatedAt string
	var closedAt *string
	err := rows.Scan(&b.ID, &b.ProductID, &branchType, &b.Name, &b.DisplayName,
		&b.SourceBranchLineID, &b.ForkedFromSnapshotID,
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
