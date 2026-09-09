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

-- branch_line: fork now points to a main VERSION, not a snapshot
CREATE TABLE IF NOT EXISTS branch_line (
    id TEXT PRIMARY KEY,
    product_id TEXT NOT NULL REFERENCES product(id) ON DELETE CASCADE,
    type TEXT NOT NULL CHECK(type IN ('MAIN', 'RELEASE')),
    name TEXT NOT NULL,
    display_name TEXT NOT NULL DEFAULT '',
    source_branch_line_id TEXT REFERENCES branch_line(id),
    forked_from_version_id TEXT REFERENCES version(id),
    status TEXT NOT NULL DEFAULT 'active'
        CHECK(status IN ('active','maintenance','security_only','end_of_support','closed')),
    created_at TEXT NOT NULL,
    updated_at TEXT NOT NULL,
    closed_at TEXT,
    UNIQUE(product_id, name)
);

-- version: replaces snapshot + branch_current_state (git-like commit on a branch line)
CREATE TABLE IF NOT EXISTS version (
    id TEXT PRIMARY KEY,
    branch_line_id TEXT NOT NULL REFERENCES branch_line(id) ON DELETE CASCADE,
    version_string TEXT NOT NULL,
    status TEXT NOT NULL DEFAULT 'incomplete'
        CHECK(status IN ('incomplete','finalized')),
    parent_version_id TEXT REFERENCES version(id),
    forked_from_version_id TEXT REFERENCES version(id),
    location TEXT,
    customer TEXT,
    release_date TEXT,
    created_at TEXT NOT NULL,
    updated_at TEXT NOT NULL,
    UNIQUE(branch_line_id, version_string)
);
CREATE INDEX IF NOT EXISTS idx_version_branch ON version(branch_line_id);
CREATE INDEX IF NOT EXISTS idx_version_parent ON version(parent_version_id);
CREATE INDEX IF NOT EXISTS idx_version_fork ON version(forked_from_version_id);

-- role-tagged DT project bindings (many per version)
CREATE TABLE IF NOT EXISTS version_dt_project (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    version_id TEXT NOT NULL REFERENCES version(id) ON DELETE CASCADE,
    role TEXT NOT NULL CHECK(role IN ('ROOT','PROFILE','SUB')),
    dt_project_uuid TEXT,
    bom_serial_number TEXT,
    bom_version INTEGER,
    bom_sha256 TEXT,
    source_revision TEXT,
    label TEXT,
    created_at TEXT NOT NULL,
    updated_at TEXT NOT NULL
);
CREATE INDEX IF NOT EXISTS idx_vdp_version ON version_dt_project(version_id);
`
