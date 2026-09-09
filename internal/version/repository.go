package version

import (
	"context"
	"database/sql"
	"time"
)

// SQLiteRepository implements Repository for SQLite.
type SQLiteRepository struct {
	db *sql.DB
}

func NewSQLiteRepository(db *sql.DB) *SQLiteRepository {
	return &SQLiteRepository{db: db}
}

const versionColumns = `id, branch_line_id, version_string, status, parent_version_id, forked_from_version_id, location, customer, release_date, created_at, updated_at`

func (r *SQLiteRepository) Create(ctx context.Context, v *Version) error {
	_, err := r.db.ExecContext(ctx,
		`INSERT INTO version (`+versionColumns+`)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		v.ID, v.BranchLineID, v.VersionString, string(v.Status),
		v.ParentVersionID, v.ForkedFromVersionID, v.Location, v.Customer,
		formatTimePtr(v.ReleaseDate), v.CreatedAt.Format(time.RFC3339), v.UpdatedAt.Format(time.RFC3339))
	return err
}

func (r *SQLiteRepository) FindByID(ctx context.Context, id string) (*Version, error) {
	row := r.db.QueryRowContext(ctx,
		`SELECT `+versionColumns+` FROM version WHERE id = ?`, id)
	return scanVersion(row)
}

// FindByBranchID lists versions ordered by release datetime (nulls last), then created_at.
func (r *SQLiteRepository) FindByBranchID(ctx context.Context, branchID string, opts ListOptions) ([]*Version, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT `+versionColumns+` FROM version WHERE branch_line_id = ?
		 ORDER BY (release_date IS NULL), release_date, created_at
		 LIMIT ? OFFSET ?`,
		branchID, opts.Limit, opts.Offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var versions []*Version
	for rows.Next() {
		v, err := scanVersionRows(rows)
		if err != nil {
			return nil, err
		}
		versions = append(versions, v)
	}
	return versions, rows.Err()
}

// FindLatestByBranchID returns the newest version by release datetime, then created_at.
func (r *SQLiteRepository) FindLatestByBranchID(ctx context.Context, branchID string) (*Version, error) {
	row := r.db.QueryRowContext(ctx,
		`SELECT `+versionColumns+` FROM version WHERE branch_line_id = ?
		 ORDER BY (release_date IS NULL), release_date DESC, created_at DESC
		 LIMIT 1`, branchID)
	return scanVersion(row)
}

func (r *SQLiteRepository) Update(ctx context.Context, v *Version) error {
	_, err := r.db.ExecContext(ctx,
		`UPDATE version SET version_string = ?, status = ?, parent_version_id = ?,
		   forked_from_version_id = ?, location = ?, customer = ?, release_date = ?, updated_at = ?
		 WHERE id = ?`,
		v.VersionString, string(v.Status), v.ParentVersionID, v.ForkedFromVersionID,
		v.Location, v.Customer, formatTimePtr(v.ReleaseDate),
		v.UpdatedAt.Format(time.RFC3339), v.ID)
	return err
}

func (r *SQLiteRepository) Delete(ctx context.Context, id string) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM version WHERE id = ?`, id)
	return err
}

// --- version_dt_project ---

const projectColumns = `id, version_id, role, dt_project_uuid, bom_serial_number, bom_version, bom_sha256, source_revision, label, created_at, updated_at`

func (r *SQLiteRepository) ListProjects(ctx context.Context, versionID string) ([]*DTProject, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT `+projectColumns+` FROM version_dt_project WHERE version_id = ? ORDER BY id`, versionID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var projects []*DTProject
	for rows.Next() {
		p, err := scanProjectRows(rows)
		if err != nil {
			return nil, err
		}
		projects = append(projects, p)
	}
	return projects, rows.Err()
}

func (r *SQLiteRepository) FindProject(ctx context.Context, versionID string, bindingID int64) (*DTProject, error) {
	row := r.db.QueryRowContext(ctx,
		`SELECT `+projectColumns+` FROM version_dt_project WHERE version_id = ? AND id = ?`,
		versionID, bindingID)
	return scanProject(row)
}

func (r *SQLiteRepository) FindProjectByRole(ctx context.Context, versionID string, role Role) (*DTProject, error) {
	row := r.db.QueryRowContext(ctx,
		`SELECT `+projectColumns+` FROM version_dt_project WHERE version_id = ? AND role = ? LIMIT 1`,
		versionID, string(role))
	return scanProject(row)
}

func (r *SQLiteRepository) CreateProject(ctx context.Context, p *DTProject) error {
	res, err := r.db.ExecContext(ctx,
		`INSERT INTO version_dt_project (version_id, role, dt_project_uuid, bom_serial_number, bom_version, bom_sha256, source_revision, label, created_at, updated_at)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		p.VersionID, string(p.Role), p.DTProjectUUID, p.BOMSerial, p.BOMVersion,
		p.BOMSHA256, p.SourceRevision, p.Label,
		p.CreatedAt.Format(time.RFC3339), p.UpdatedAt.Format(time.RFC3339))
	if err != nil {
		return err
	}
	id, err := res.LastInsertId()
	if err != nil {
		return err
	}
	p.ID = id
	return nil
}

func (r *SQLiteRepository) UpdateProject(ctx context.Context, p *DTProject) error {
	_, err := r.db.ExecContext(ctx,
		`UPDATE version_dt_project SET role = ?, dt_project_uuid = ?, bom_serial_number = ?,
		   bom_version = ?, bom_sha256 = ?, source_revision = ?, label = ?, updated_at = ?
		 WHERE id = ?`,
		string(p.Role), p.DTProjectUUID, p.BOMSerial, p.BOMVersion, p.BOMSHA256,
		p.SourceRevision, p.Label, p.UpdatedAt.Format(time.RFC3339), p.ID)
	return err
}

func (r *SQLiteRepository) DeleteProject(ctx context.Context, versionID string, bindingID int64) error {
	_, err := r.db.ExecContext(ctx,
		`DELETE FROM version_dt_project WHERE version_id = ? AND id = ?`, versionID, bindingID)
	return err
}

// --- scan helpers ---

func formatTimePtr(t *time.Time) *string {
	if t == nil {
		return nil
	}
	s := t.Format(time.RFC3339)
	return &s
}

func parseTimePtr(s *string) *time.Time {
	if s == nil || *s == "" {
		return nil
	}
	t, err := time.Parse(time.RFC3339, *s)
	if err != nil {
		return nil
	}
	return &t
}

type scanner interface {
	Scan(dest ...any) error
}

func scanVersion(row *sql.Row) (*Version, error) {
	v, err := scanVersionInto(row)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return v, err
}

func scanVersionRows(rows *sql.Rows) (*Version, error) {
	return scanVersionInto(rows)
}

func scanVersionInto(s scanner) (*Version, error) {
	var v Version
	var status, createdAt, updatedAt string
	var releaseDate *string
	err := s.Scan(&v.ID, &v.BranchLineID, &v.VersionString, &status,
		&v.ParentVersionID, &v.ForkedFromVersionID, &v.Location, &v.Customer,
		&releaseDate, &createdAt, &updatedAt)
	if err != nil {
		return nil, err
	}
	v.Status = Status(status)
	v.CreatedAt, _ = time.Parse(time.RFC3339, createdAt)
	v.UpdatedAt, _ = time.Parse(time.RFC3339, updatedAt)
	v.ReleaseDate = parseTimePtr(releaseDate)
	return &v, nil
}

func scanProject(row *sql.Row) (*DTProject, error) {
	p, err := scanProjectInto(row)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return p, err
}

func scanProjectRows(rows *sql.Rows) (*DTProject, error) {
	return scanProjectInto(rows)
}

func scanProjectInto(s scanner) (*DTProject, error) {
	var p DTProject
	var role, createdAt, updatedAt string
	err := s.Scan(&p.ID, &p.VersionID, &role, &p.DTProjectUUID, &p.BOMSerial,
		&p.BOMVersion, &p.BOMSHA256, &p.SourceRevision, &p.Label, &createdAt, &updatedAt)
	if err != nil {
		return nil, err
	}
	p.Role = Role(role)
	p.CreatedAt, _ = time.Parse(time.RFC3339, createdAt)
	p.UpdatedAt, _ = time.Parse(time.RFC3339, updatedAt)
	return &p, nil
}
