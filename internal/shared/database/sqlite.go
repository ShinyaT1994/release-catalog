package database

import "database/sql"

// Migrate runs all SQLite migrations
func Migrate(db *sql.DB) error {
	_, err := db.Exec(migrationSQL)
	return err
}

const migrationSQL = `
CREATE TABLE IF NOT EXISTS product (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL UNIQUE,
    display_name TEXT NOT NULL DEFAULT '',
    description TEXT NOT NULL DEFAULT '',
    created_at TEXT NOT NULL,
    updated_at TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS branch_line (
    id TEXT PRIMARY KEY,
    product_id TEXT NOT NULL REFERENCES product(id) ON DELETE CASCADE,
    type TEXT NOT NULL CHECK(type IN ('MAIN', 'RELEASE')),
    name TEXT NOT NULL,
    display_name TEXT NOT NULL DEFAULT '',
    source_branch_line_id TEXT REFERENCES branch_line(id),
    forked_from_snapshot_id TEXT,
    status TEXT NOT NULL DEFAULT 'active'
        CHECK(status IN ('active','maintenance','security_only','end_of_support','closed')),
    created_at TEXT NOT NULL,
    updated_at TEXT NOT NULL,
    closed_at TEXT,
    UNIQUE(product_id, name)
);

CREATE TABLE IF NOT EXISTS branch_current_state (
    branch_line_id TEXT PRIMARY KEY REFERENCES branch_line(id) ON DELETE CASCADE,
    root_dt_project_uuid TEXT,
    root_bom_serial_number TEXT,
    root_bom_version INTEGER,
    root_bom_sha256 TEXT,
    source_revision TEXT,
    updated_at TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS snapshot (
    id TEXT PRIMARY KEY,
    branch_line_id TEXT NOT NULL REFERENCES branch_line(id) ON DELETE CASCADE,
    snapshot_type TEXT NOT NULL CHECK(snapshot_type IN ('MAIN_SNAPSHOT', 'RELEASE')),
    version TEXT NOT NULL,
    status TEXT NOT NULL DEFAULT 'draft'
        CHECK(status IN ('draft','testing','approved','released','deprecated','end_of_support')),
    root_dt_project_uuid TEXT,
    root_bom_serial_number TEXT,
    root_bom_version INTEGER,
    root_bom_sha256 TEXT,
    source_revision TEXT,
    created_at TEXT NOT NULL,
    released_at TEXT
);

CREATE TABLE IF NOT EXISTS bom_link_index (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    source_dt_project_uuid TEXT NOT NULL,
    source_bom_serial TEXT,
    source_bom_version INTEGER,
    source_bom_ref TEXT,
    target_serial TEXT,
    target_bom_version INTEGER,
    target_bom_ref TEXT,
    target_dt_project_uuid TEXT,
    resolution_status TEXT NOT NULL DEFAULT 'pending'
        CHECK(resolution_status IN ('resolved','missing_project','missing_bom','missing_bom_ref','invalid','pending')),
    updated_at TEXT NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_bom_link_source ON bom_link_index(source_dt_project_uuid);
CREATE INDEX IF NOT EXISTS idx_bom_link_target ON bom_link_index(target_dt_project_uuid);
`
