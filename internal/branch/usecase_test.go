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

// mockProductFinder always returns true.
type mockProductFinder struct{}

func (m *mockProductFinder) FindByID(ctx context.Context, id string) (bool, error) {
	return true, nil
}

// mockVersionFinder returns a configurable version for fork validation.
type mockVersionFinder struct {
	info *branch.VersionInfo
}

func (m *mockVersionFinder) FindByID(ctx context.Context, id string) (*branch.VersionInfo, error) {
	if m.info != nil && m.info.ID == id {
		return m.info, nil
	}
	return nil, nil
}

// mockVersionCreator records calls to CreateFirstVersion.
type mockVersionCreator struct {
	called   bool
	branchID string
	input    branch.CreateFirstVersionInput
}

func (m *mockVersionCreator) CreateFirstVersion(ctx context.Context, branchID string, input branch.CreateFirstVersionInput) error {
	m.called = true
	m.branchID = branchID
	m.input = input
	return nil
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
	svc := branch.NewService(repo, &mockProductFinder{}, &mockVersionFinder{}, nil)

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
	assert.Nil(t, b.ForkedFromVersionID)
}

func TestCreateReleaseLine_ForkFromMainVersion(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	productID := seedProduct(t, db)
	mainBranch := seedMainBranch(t, db, productID)

	// Seed a main version to fork from.
	versionID := uuid.New().String()
	now := time.Now().UTC().Format(time.RFC3339)
	_, err := db.Exec(`INSERT INTO version (id, branch_line_id, version_string, status, created_at, updated_at) VALUES (?, ?, ?, 'finalized', ?, ?)`,
		versionID, mainBranch.ID, "1.0", now, now)
	require.NoError(t, err)

	repo := branch.NewSQLiteRepository(db)
	vf := &mockVersionFinder{info: &branch.VersionInfo{ID: versionID, BranchLineID: mainBranch.ID, VersionString: "1.0"}}
	vc := &mockVersionCreator{}
	svc := branch.NewService(repo, &mockProductFinder{}, vf, vc)

	b, err := svc.CreateReleaseLine(context.Background(), productID, branch.CreateReleaseLineInput{
		Name:                "release/1.x",
		ForkedFromVersionID: &versionID,
	})
	require.NoError(t, err)
	require.NotNil(t, b.ForkedFromVersionID)
	assert.Equal(t, versionID, *b.ForkedFromVersionID)
	assert.Equal(t, &mainBranch.ID, b.SourceBranchLineID)

	// The first version must be auto-created on the new release branch, forked
	// from the main version, inheriting its version string.
	assert.True(t, vc.called, "expected first version to be auto-created")
	assert.Equal(t, b.ID, vc.branchID)
	require.NotNil(t, vc.input.ForkedFromVersionID)
	assert.Equal(t, versionID, *vc.input.ForkedFromVersionID)
	assert.Equal(t, "1.0", vc.input.VersionString)
}

func TestCreateReleaseLine_ForkVersionNotFound(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	productID := seedProduct(t, db)
	seedMainBranch(t, db, productID)

	repo := branch.NewSQLiteRepository(db)
	svc := branch.NewService(repo, &mockProductFinder{}, &mockVersionFinder{}, nil)

	missing := uuid.New().String()
	_, err := svc.CreateReleaseLine(context.Background(), productID, branch.CreateReleaseLineInput{
		Name:                "release/1.x",
		ForkedFromVersionID: &missing,
	})
	require.Error(t, err)
}

func TestUpdateBranchStatus(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	productID := seedProduct(t, db)
	seedMainBranch(t, db, productID)

	repo := branch.NewSQLiteRepository(db)
	svc := branch.NewService(repo, &mockProductFinder{}, &mockVersionFinder{}, nil)
	ctx := context.Background()

	b, _ := svc.CreateReleaseLine(ctx, productID, branch.CreateReleaseLineInput{Name: "release/1.x"})

	newStatus := branch.StatusSecurityOnly
	updated, err := svc.Update(ctx, b.ID, branch.UpdateBranchInput{Status: &newStatus})
	require.NoError(t, err)
	assert.Equal(t, branch.StatusSecurityOnly, updated.Status)
}
