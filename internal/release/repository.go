package release

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

func (r *SQLiteRepository) Create(ctx context.Context, s *Snapshot) error {
	var releasedAt *string
	if s.ReleasedAt != nil {
		ra := s.ReleasedAt.Format(time.RFC3339)
		releasedAt = &ra
	}
	_, err := r.db.ExecContext(ctx,
		`INSERT INTO snapshot (id, branch_line_id, snapshot_type, version, status, root_dt_project_uuid, root_bom_serial_number, root_bom_version, root_bom_sha256, source_revision, created_at, released_at)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		s.ID, s.BranchLineID, string(s.SnapshotType), s.Version, string(s.Status),
		s.RootDTProjectUUID, s.RootBOMSerial, s.RootBOMVersion, s.RootBOMSHA256,
		s.SourceRevision, s.CreatedAt.Format(time.RFC3339), releasedAt)
	return err
}

func (r *SQLiteRepository) FindByID(ctx context.Context, id string) (*Snapshot, error) {
	row := r.db.QueryRowContext(ctx,
		`SELECT id, branch_line_id, snapshot_type, version, status, root_dt_project_uuid, root_bom_serial_number, root_bom_version, root_bom_sha256, source_revision, created_at, released_at
		 FROM snapshot WHERE id = ?`, id)
	return scanSnapshot(row)
}

func (r *SQLiteRepository) FindByBranchID(ctx context.Context, branchID string, opts ListOptions) ([]*Snapshot, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT id, branch_line_id, snapshot_type, version, status, root_dt_project_uuid, root_bom_serial_number, root_bom_version, root_bom_sha256, source_revision, created_at, released_at
		 FROM snapshot WHERE branch_line_id = ? ORDER BY created_at DESC LIMIT ? OFFSET ?`,
		branchID, opts.Limit, opts.Offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var snapshots []*Snapshot
	for rows.Next() {
		s, err := scanSnapshotRows(rows)
		if err != nil {
			return nil, err
		}
		snapshots = append(snapshots, s)
	}
	return snapshots, rows.Err()
}

func scanSnapshot(row *sql.Row) (*Snapshot, error) {
	var s Snapshot
	var snapshotType, status, createdAt string
	var releasedAt *string
	err := row.Scan(&s.ID, &s.BranchLineID, &snapshotType, &s.Version, &status,
		&s.RootDTProjectUUID, &s.RootBOMSerial, &s.RootBOMVersion, &s.RootBOMSHA256,
		&s.SourceRevision, &createdAt, &releasedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	s.SnapshotType = SnapshotType(snapshotType)
	s.Status = Status(status)
	s.CreatedAt, _ = time.Parse(time.RFC3339, createdAt)
	if releasedAt != nil {
		t, _ := time.Parse(time.RFC3339, *releasedAt)
		s.ReleasedAt = &t
	}
	return &s, nil
}

func scanSnapshotRows(rows *sql.Rows) (*Snapshot, error) {
	var s Snapshot
	var snapshotType, status, createdAt string
	var releasedAt *string
	err := rows.Scan(&s.ID, &s.BranchLineID, &snapshotType, &s.Version, &status,
		&s.RootDTProjectUUID, &s.RootBOMSerial, &s.RootBOMVersion, &s.RootBOMSHA256,
		&s.SourceRevision, &createdAt, &releasedAt)
	if err != nil {
		return nil, err
	}
	s.SnapshotType = SnapshotType(snapshotType)
	s.Status = Status(status)
	s.CreatedAt, _ = time.Parse(time.RFC3339, createdAt)
	if releasedAt != nil {
		t, _ := time.Parse(time.RFC3339, *releasedAt)
		s.ReleasedAt = &t
	}
	return &s, nil
}
