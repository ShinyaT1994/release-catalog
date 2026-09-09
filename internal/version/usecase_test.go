package version_test

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/ShinyaT1994/release-catalog/internal/shared/database"
	"github.com/ShinyaT1994/release-catalog/internal/version"
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

// mockBranchFinder returns branch info for the seeded branch id.
type mockBranchFinder struct {
	id       string
	typeName string
}

func (m *mockBranchFinder) FindByID(ctx context.Context, id string) (*version.BranchInfo, error) {
	if id == m.id {
		return &version.BranchInfo{ID: m.id, Type: m.typeName}, nil
	}
	return nil, nil
}

func seedBranch(t *testing.T, db *sql.DB, typeName string) string {
	pid := uuid.New().String()
	bid := uuid.New().String()
	now := time.Now().UTC().Format(time.RFC3339)
	_, err := db.Exec(`INSERT INTO product (id, name, display_name, description, created_at, updated_at) VALUES (?, ?, '', '', ?, ?)`,
		pid, "p-"+pid[:8], now, now)
	require.NoError(t, err)
	_, err = db.Exec(`INSERT INTO branch_line (id, product_id, type, name, display_name, status, created_at, updated_at) VALUES (?, ?, ?, ?, '', 'active', ?, ?)`,
		bid, pid, typeName, "b-"+bid[:8], now, now)
	require.NoError(t, err)
	return bid
}

func newSvc(db *sql.DB, branchID, typeName string) *version.Service {
	repo := version.NewSQLiteRepository(db)
	return version.NewService(repo, &mockBranchFinder{id: branchID, typeName: typeName})
}

func TestCreateVersion_PresetMinimal(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()
	bid := seedBranch(t, db, "MAIN")
	svc := newSvc(db, bid, "MAIN")

	v, err := svc.Create(context.Background(), bid, version.CreateInput{VersionString: "1.0"})
	require.NoError(t, err)
	assert.Equal(t, "1.0", v.VersionString)
	assert.Equal(t, version.StatusIncomplete, v.Status)
	assert.Nil(t, v.ParentVersionID)
	assert.Empty(t, v.Projects)
}

func TestCreateVersion_RequiresVersionString(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()
	bid := seedBranch(t, db, "MAIN")
	svc := newSvc(db, bid, "MAIN")

	_, err := svc.Create(context.Background(), bid, version.CreateInput{})
	require.Error(t, err)
}

func TestCreateVersion_ParentIsLatest(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()
	bid := seedBranch(t, db, "MAIN")
	svc := newSvc(db, bid, "MAIN")
	ctx := context.Background()

	v1, err := svc.Create(ctx, bid, version.CreateInput{VersionString: "1.0"})
	require.NoError(t, err)
	// ensure distinct release-date/created ordering: set release date on v1
	rd := time.Now().UTC().Format(time.RFC3339)
	_, err = svc.Update(ctx, v1.ID, version.UpdateInput{ReleaseDate: &rd})
	require.NoError(t, err)

	v2, err := svc.Create(ctx, bid, version.CreateInput{VersionString: "1.1"})
	require.NoError(t, err)
	require.NotNil(t, v2.ParentVersionID)
	assert.Equal(t, v1.ID, *v2.ParentVersionID)
}

func TestSetRoot_EnforcesSingleRoot(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()
	bid := seedBranch(t, db, "MAIN")
	svc := newSvc(db, bid, "MAIN")
	ctx := context.Background()

	v, err := svc.Create(ctx, bid, version.CreateInput{VersionString: "1.0"})
	require.NoError(t, err)

	uuid1 := "11111111-1111-1111-1111-111111111111"
	uuid2 := "22222222-2222-2222-2222-222222222222"
	_, err = svc.SetRoot(ctx, v.ID, version.SetProjectInput{DTProjectUUID: &uuid1})
	require.NoError(t, err)
	_, err = svc.SetRoot(ctx, v.ID, version.SetProjectInput{DTProjectUUID: &uuid2})
	require.NoError(t, err)

	got, err := svc.GetByID(ctx, v.ID)
	require.NoError(t, err)
	roots := 0
	for _, p := range got.Projects {
		if p.Role == version.RoleRoot {
			roots++
			assert.Equal(t, uuid2, *p.DTProjectUUID)
		}
	}
	assert.Equal(t, 1, roots, "must have exactly one ROOT after two SetRoot calls")
}

func TestAddProject_RejectsRoot(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()
	bid := seedBranch(t, db, "MAIN")
	svc := newSvc(db, bid, "MAIN")
	ctx := context.Background()

	v, _ := svc.Create(ctx, bid, version.CreateInput{VersionString: "1.0"})
	_, err := svc.AddProject(ctx, v.ID, version.AddProjectInput{Role: version.RoleRoot})
	require.Error(t, err)
}

func TestAddProject_ProfileAndSub(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()
	bid := seedBranch(t, db, "RELEASE")
	svc := newSvc(db, bid, "RELEASE")
	ctx := context.Background()

	v, _ := svc.Create(ctx, bid, version.CreateInput{VersionString: "1.0"})
	_, err := svc.AddProject(ctx, v.ID, version.AddProjectInput{Role: version.RoleProfile})
	require.NoError(t, err)
	_, err = svc.AddProject(ctx, v.ID, version.AddProjectInput{Role: version.RoleSub})
	require.NoError(t, err)

	got, _ := svc.GetByID(ctx, v.ID)
	assert.Len(t, got.Projects, 2)
}

func TestForkDefaultsRootFromMainVersion(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()
	ctx := context.Background()

	// Main branch + version with a ROOT.
	mainBID := seedBranch(t, db, "MAIN")
	mainSvc := newSvc(db, mainBID, "MAIN")
	mainV, err := mainSvc.Create(ctx, mainBID, version.CreateInput{VersionString: "1.0"})
	require.NoError(t, err)
	rootUUID := "aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa"
	_, err = mainSvc.SetRoot(ctx, mainV.ID, version.SetProjectInput{DTProjectUUID: &rootUUID})
	require.NoError(t, err)

	// Release branch's first version forks from the main version.
	relBID := seedBranch(t, db, "RELEASE")
	relSvc := newSvc(db, relBID, "RELEASE")
	relV, err := relSvc.Create(ctx, relBID, version.CreateInput{
		VersionString:       "1.0-1",
		ForkedFromVersionID: &mainV.ID,
	})
	require.NoError(t, err)
	require.NotNil(t, relV.ForkedFromVersionID)

	// The ROOT should be defaulted from the forked main version.
	var rootProjects int
	for _, p := range relV.Projects {
		if p.Role == version.RoleRoot {
			rootProjects++
			require.NotNil(t, p.DTProjectUUID)
			assert.Equal(t, rootUUID, *p.DTProjectUUID)
		}
	}
	assert.Equal(t, 1, rootProjects)
}

func TestListOrderedByReleaseDate(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()
	bid := seedBranch(t, db, "MAIN")
	svc := newSvc(db, bid, "MAIN")
	ctx := context.Background()

	vA, _ := svc.Create(ctx, bid, version.CreateInput{VersionString: "A"})
	vB, _ := svc.Create(ctx, bid, version.CreateInput{VersionString: "B"})

	// vB has an earlier release date than vA -> B should come first.
	early := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC).Format(time.RFC3339)
	late := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC).Format(time.RFC3339)
	_, err := svc.Update(ctx, vA.ID, version.UpdateInput{ReleaseDate: &late})
	require.NoError(t, err)
	_, err = svc.Update(ctx, vB.ID, version.UpdateInput{ReleaseDate: &early})
	require.NoError(t, err)

	list, err := svc.ListByBranchID(ctx, bid, version.DefaultListOptions())
	require.NoError(t, err)
	require.Len(t, list, 2)
	assert.Equal(t, vB.ID, list[0].ID)
	assert.Equal(t, vA.ID, list[1].ID)
}

func TestUpdateSoftImmutability(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()
	bid := seedBranch(t, db, "RELEASE")
	svc := newSvc(db, bid, "RELEASE")
	ctx := context.Background()

	v, _ := svc.Create(ctx, bid, version.CreateInput{VersionString: "1.0"})
	newStr := "1.0-final"
	fin := version.StatusFinalized
	cust := "ACME"
	updated, err := svc.Update(ctx, v.ID, version.UpdateInput{
		VersionString: &newStr,
		Status:        &fin,
		Customer:      &cust,
	})
	require.NoError(t, err)
	assert.Equal(t, "1.0-final", updated.VersionString)
	assert.Equal(t, version.StatusFinalized, updated.Status)
	require.NotNil(t, updated.Customer)
	assert.Equal(t, "ACME", *updated.Customer)
}
