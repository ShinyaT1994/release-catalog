package branch_test

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/ShinyaT1994/release-catalog/internal/branch"
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

// mockProductFinder always returns true
type mockProductFinder struct{}

func (m *mockProductFinder) FindByID(ctx context.Context, id string) (bool, error) {
	return true, nil
}

type mockSnapshotFinder struct{}

func (m *mockSnapshotFinder) FindByID(ctx context.Context, id string) (*branch.SnapshotInfo, error) {
	return nil, nil
}

func seedMainBranch(t *testing.T, db *sql.DB, productID string) *branch.BranchLine {
	repo := branch.NewSQLiteRepository(db)
	now := time.Now().UTC()
	b := &branch.BranchLine{
		ID: uuid.New().String(), ProductID: productID,
		Type: branch.TypeMain, Name: "main", DisplayName: "Main",
		Status: branch.StatusActive, CreatedAt: now, UpdatedAt: now,
	}
	require.NoError(t, repo.Create(context.Background(), b))
	return b
}

func seedProduct(t *testing.T, db *sql.DB) string {
	now := time.Now().UTC().Format(time.RFC3339)
	id := uuid.New().String()
	_, err := db.Exec(`INSERT INTO product (id, name, display_name, description, created_at, updated_at) VALUES (?, ?, '', '', ?, ?)`,
		id, "test-product-"+id[:8], now, now)
	require.NoError(t, err)
	return id
}

func TestCreateReleaseLine(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	productID := seedProduct(t, db)
	seedMainBranch(t, db, productID)

	repo := branch.NewSQLiteRepository(db)
	csRepo := branch.NewSQLiteCurrentStateRepository(db)
	svc := branch.NewService(repo, csRepo, &mockProductFinder{}, &mockSnapshotFinder{})

	ctx := context.Background()
	b, err := svc.CreateReleaseLine(ctx, productID, branch.CreateReleaseLineInput{
		Name:        "release/2.x",
		DisplayName: "Release 2.x",
	})

	require.NoError(t, err)
	assert.NotEmpty(t, b.ID)
	assert.Equal(t, branch.TypeRelease, b.Type)
	assert.Equal(t, "release/2.x", b.Name)
	assert.Equal(t, branch.StatusActive, b.Status)
}

func TestReleaseLine_IndependentFromMain(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	productID := seedProduct(t, db)
	mainBranch := seedMainBranch(t, db, productID)

	repo := branch.NewSQLiteRepository(db)
	csRepo := branch.NewSQLiteCurrentStateRepository(db)
	svc := branch.NewService(repo, csRepo, &mockProductFinder{}, &mockSnapshotFinder{})
	ctx := context.Background()

	// Set main current state
	mainProjectUUID := "00000000-0000-0000-0000-000000000001"
	_, err := svc.UpdateCurrentState(ctx, mainBranch.ID, branch.UpdateCurrentStateInput{
		RootDTProjectUUID: &mainProjectUUID,
	})
	require.NoError(t, err)

	// Create release line (forks from main current)
	releaseBranch, err := svc.CreateReleaseLine(ctx, productID, branch.CreateReleaseLineInput{
		Name: "release/1.x",
	})
	require.NoError(t, err)

	// Update main - should NOT affect release
	newMainUUID := "00000000-0000-0000-0000-000000000099"
	_, err = svc.UpdateCurrentState(ctx, mainBranch.ID, branch.UpdateCurrentStateInput{
		RootDTProjectUUID: &newMainUUID,
	})
	require.NoError(t, err)

	// Release branch current should still point to original
	releaseCS, err := svc.GetCurrentState(ctx, releaseBranch.ID)
	require.NoError(t, err)
	assert.Equal(t, &mainProjectUUID, releaseCS.RootDTProjectUUID)

	// Main current should have new value
	mainCS, err := svc.GetCurrentState(ctx, mainBranch.ID)
	require.NoError(t, err)
	assert.Equal(t, &newMainUUID, mainCS.RootDTProjectUUID)
}

func TestUpdateBranchStatus(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	productID := seedProduct(t, db)
	seedMainBranch(t, db, productID)

	repo := branch.NewSQLiteRepository(db)
	csRepo := branch.NewSQLiteCurrentStateRepository(db)
	svc := branch.NewService(repo, csRepo, &mockProductFinder{}, &mockSnapshotFinder{})
	ctx := context.Background()

	b, _ := svc.CreateReleaseLine(ctx, productID, branch.CreateReleaseLineInput{Name: "release/1.x"})

	newStatus := branch.StatusSecurityOnly
	updated, err := svc.Update(ctx, b.ID, branch.UpdateBranchInput{Status: &newStatus})
	require.NoError(t, err)
	assert.Equal(t, branch.StatusSecurityOnly, updated.Status)
}
