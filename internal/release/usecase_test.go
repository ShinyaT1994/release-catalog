package release_test

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/ShinyaT1994/release-catalog/internal/branch"
	"github.com/ShinyaT1994/release-catalog/internal/release"
	"github.com/ShinyaT1994/release-catalog/internal/shared/database"
	"github.com/google/uuid"
	_ "github.com/mattn/go-sqlite3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupTestDB(t *testing.T) *sql.DB {
	db, err := sql.Open("sqlite3", ":memory:?_foreign_keys=on")
	require.NoError(t, err)
	require.NoError(t, database.Migrate(db))
	return db
}

type mockBranchFinder struct {
	branches map[string]*release.BranchInfo
}

func (m *mockBranchFinder) FindByID(ctx context.Context, id string) (*release.BranchInfo, error) {
	return m.branches[id], nil
}

type mockCSFinder struct {
	states map[string]*release.CurrentStateInfo
}

func (m *mockCSFinder) FindByBranchID(ctx context.Context, branchID string) (*release.CurrentStateInfo, error) {
	return m.states[branchID], nil
}

func seedForRelease(t *testing.T, db *sql.DB) (string, string) {
	now := time.Now().UTC().Format(time.RFC3339)
	productID := uuid.New().String()
	_, err := db.Exec(`INSERT INTO product (id, name, display_name, description, created_at, updated_at) VALUES (?, ?, '', '', ?, ?)`,
		productID, "test-"+productID[:8], now, now)
	require.NoError(t, err)

	branchID := uuid.New().String()
	_, err = db.Exec(`INSERT INTO branch_line (id, product_id, type, name, display_name, status, created_at, updated_at) VALUES (?, ?, 'RELEASE', 'release/1.x', 'Release 1.x', 'active', ?, ?)`,
		branchID, productID, now, now)
	require.NoError(t, err)

	return productID, branchID
}

func seedMainBranch(t *testing.T, db *sql.DB, productID string) string {
	now := time.Now().UTC().Format(time.RFC3339)
	branchID := uuid.New().String()
	_, err := db.Exec(`INSERT INTO branch_line (id, product_id, type, name, display_name, status, created_at, updated_at) VALUES (?, ?, 'MAIN', 'main', 'Main', 'active', ?, ?)`,
		branchID, productID, now, now)
	require.NoError(t, err)
	return branchID
}

func TestCreateRelease(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	_, branchID := seedForRelease(t, db)

	repo := release.NewSQLiteRepository(db)
	projectUUID := "00000000-0000-0000-0000-000000000001"
	svc := release.NewService(repo,
		&mockBranchFinder{branches: map[string]*release.BranchInfo{branchID: {ID: branchID, Type: "RELEASE"}}},
		&mockCSFinder{states: map[string]*release.CurrentStateInfo{branchID: {RootDTProjectUUID: &projectUUID}}},
	)

	snap, err := svc.CreateRelease(context.Background(), branchID, release.CreateReleaseInput{Version: "1.0.0"})
	require.NoError(t, err)
	assert.Equal(t, "1.0.0", snap.Version)
	assert.Equal(t, release.TypeRelease, snap.SnapshotType)
	assert.Equal(t, release.StatusReleased, snap.Status)
	assert.NotNil(t, snap.ReleasedAt)
	assert.Equal(t, &projectUUID, snap.RootDTProjectUUID)
}

func TestCreateRelease_OnMainBranch_Fails(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	now := time.Now().UTC().Format(time.RFC3339)
	productID := uuid.New().String()
	db.Exec(`INSERT INTO product (id, name, display_name, description, created_at, updated_at) VALUES (?, 'test', '', '', ?, ?)`,
		productID, now, now)
	mainBranchID := seedMainBranch(t, db, productID)

	repo := release.NewSQLiteRepository(db)
	svc := release.NewService(repo,
		&mockBranchFinder{branches: map[string]*release.BranchInfo{mainBranchID: {ID: mainBranchID, Type: "MAIN"}}},
		&mockCSFinder{states: map[string]*release.CurrentStateInfo{}},
	)

	_, err := svc.CreateRelease(context.Background(), mainBranchID, release.CreateReleaseInput{Version: "1.0.0"})
	assert.Error(t, err)
}

func TestSnapshotImmutability(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	_, branchID := seedForRelease(t, db)

	repo := release.NewSQLiteRepository(db)
	projectUUID := "00000000-0000-0000-0000-000000000001"
	svc := release.NewService(repo,
		&mockBranchFinder{branches: map[string]*release.BranchInfo{branchID: {ID: branchID, Type: "RELEASE"}}},
		&mockCSFinder{states: map[string]*release.CurrentStateInfo{branchID: {RootDTProjectUUID: &projectUUID}}},
	)

	// Create first release
	snap1, err := svc.CreateRelease(context.Background(), branchID, release.CreateReleaseInput{Version: "1.0.0"})
	require.NoError(t, err)

	// Get it back - should be unchanged
	got, err := svc.GetByID(context.Background(), snap1.ID)
	require.NoError(t, err)
	assert.Equal(t, snap1.Version, got.Version)
	assert.Equal(t, snap1.RootDTProjectUUID, got.RootDTProjectUUID)
}

func TestListReleases(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	_, branchID := seedForRelease(t, db)
	csRepo := branch.NewSQLiteCurrentStateRepository(db)
	csRepo.Upsert(context.Background(), &branch.CurrentState{BranchLineID: branchID, UpdatedAt: time.Now().UTC()})

	repo := release.NewSQLiteRepository(db)
	svc := release.NewService(repo,
		&mockBranchFinder{branches: map[string]*release.BranchInfo{branchID: {ID: branchID, Type: "RELEASE"}}},
		&mockCSFinder{states: map[string]*release.CurrentStateInfo{branchID: {}}},
	)

	ctx := context.Background()
	svc.CreateRelease(ctx, branchID, release.CreateReleaseInput{Version: "1.0.0"})
	svc.CreateRelease(ctx, branchID, release.CreateReleaseInput{Version: "1.0.1"})

	list, err := svc.ListByBranchID(ctx, branchID, release.DefaultListOptions())
	require.NoError(t, err)
	assert.Len(t, list, 2)
}
