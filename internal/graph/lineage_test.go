package graph_test

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/ShinyaT1994/release-catalog/internal/graph"
	"github.com/ShinyaT1994/release-catalog/internal/shared/database"
	"github.com/google/uuid"
	_ "github.com/mattn/go-sqlite3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupLineageDB(t *testing.T) *sql.DB {
	db, err := sql.Open("sqlite3", ":memory:?_foreign_keys=on")
	require.NoError(t, err)
	require.NoError(t, database.Migrate(db))
	return db
}

func TestLineage_ParentAndForkEdges(t *testing.T) {
	db := setupLineageDB(t)
	defer db.Close()
	ctx := context.Background()

	now := time.Now().UTC().Format(time.RFC3339)
	pid := uuid.New().String()
	mainB := uuid.New().String()
	relB := uuid.New().String()
	_, err := db.Exec(`INSERT INTO product (id, name, created_at, updated_at) VALUES (?, 'prod', ?, ?)`, pid, now, now)
	require.NoError(t, err)
	_, err = db.Exec(`INSERT INTO branch_line (id, product_id, type, name, status, created_at, updated_at) VALUES (?, ?, 'MAIN', 'main', 'active', ?, ?)`, mainB, pid, now, now)
	require.NoError(t, err)
	_, err = db.Exec(`INSERT INTO branch_line (id, product_id, type, name, status, created_at, updated_at) VALUES (?, ?, 'RELEASE', 'rel', 'active', ?, ?)`, relB, pid, now, now)
	require.NoError(t, err)

	// main: m1 -> m2 (parent). release: r1 forked from m1.
	m1 := uuid.New().String()
	m2 := uuid.New().String()
	r1 := uuid.New().String()
	d1 := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC).Format(time.RFC3339)
	d2 := time.Date(2024, 2, 1, 0, 0, 0, 0, time.UTC).Format(time.RFC3339)
	d3 := time.Date(2024, 3, 1, 0, 0, 0, 0, time.UTC).Format(time.RFC3339)
	_, err = db.Exec(`INSERT INTO version (id, branch_line_id, version_string, status, release_date, created_at, updated_at) VALUES (?, ?, '1.0', 'finalized', ?, ?, ?)`, m1, mainB, d1, now, now)
	require.NoError(t, err)
	_, err = db.Exec(`INSERT INTO version (id, branch_line_id, version_string, status, parent_version_id, release_date, created_at, updated_at) VALUES (?, ?, '1.1', 'finalized', ?, ?, ?, ?)`, m2, mainB, m1, d3, now, now)
	require.NoError(t, err)
	_, err = db.Exec(`INSERT INTO version (id, branch_line_id, version_string, status, forked_from_version_id, release_date, created_at, updated_at) VALUES (?, ?, '1.0-1', 'finalized', ?, ?, ?, ?)`, r1, relB, m1, d2, now, now)
	require.NoError(t, err)

	src := graph.NewSQLiteLineageSource(db)
	svc := graph.NewLineageService(src)

	g, err := svc.GetProductLineage(ctx, pid, graph.AxisReleaseDate)
	require.NoError(t, err)
	assert.Len(t, g.Nodes, 3)

	var parentEdges, forkEdges int
	for _, e := range g.Edges {
		switch e.Kind {
		case "parent":
			parentEdges++
			assert.Equal(t, m1, e.SourceVersionID)
			assert.Equal(t, m2, e.TargetVersionID)
		case "fork":
			forkEdges++
			assert.Equal(t, m1, e.SourceVersionID)
			assert.Equal(t, r1, e.TargetVersionID)
		}
	}
	assert.Equal(t, 1, parentEdges)
	assert.Equal(t, 1, forkEdges)

	// Ordered by release date: m1(d1), r1(d2), m2(d3)
	assert.Equal(t, m1, g.Nodes[0].VersionID)
	assert.Equal(t, r1, g.Nodes[1].VersionID)
	assert.Equal(t, m2, g.Nodes[2].VersionID)
}

func TestLineage_ProductNotFound(t *testing.T) {
	db := setupLineageDB(t)
	defer db.Close()
	src := graph.NewSQLiteLineageSource(db)
	svc := graph.NewLineageService(src)
	_, err := svc.GetProductLineage(context.Background(), "nope", graph.AxisVersion)
	require.Error(t, err)
}

func TestTimeline_Product(t *testing.T) {
	db := setupLineageDB(t)
	defer db.Close()
	now := time.Now().UTC().Format(time.RFC3339)
	pid := uuid.New().String()
	b := uuid.New().String()
	_, _ = db.Exec(`INSERT INTO product (id, name, created_at, updated_at) VALUES (?, 'p', ?, ?)`, pid, now, now)
	_, _ = db.Exec(`INSERT INTO branch_line (id, product_id, type, name, status, created_at, updated_at) VALUES (?, ?, 'MAIN', 'main', 'active', ?, ?)`, b, pid, now, now)
	_, _ = db.Exec(`INSERT INTO version (id, branch_line_id, version_string, status, created_at, updated_at) VALUES (?, ?, '1.0', 'finalized', ?, ?)`, uuid.New().String(), b, now, now)

	src := graph.NewSQLiteLineageSource(db)
	svc := graph.NewLineageService(src)
	tl, err := svc.GetProductTimeline(context.Background(), pid)
	require.NoError(t, err)
	assert.Len(t, tl.Entries, 1)
}
