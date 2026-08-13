package product_test

import (
	"context"
	"database/sql"
	"testing"

	"github.com/ShinyaT1994/release-catalog/internal/branch"
	"github.com/ShinyaT1994/release-catalog/internal/product"
	"github.com/ShinyaT1994/release-catalog/internal/shared/database"
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

func TestCreateProduct(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	productRepo := product.NewSQLiteRepository(db)
	branchRepo := branch.NewSQLiteRepository(db)
	svc := product.NewService(productRepo, branchRepo)

	ctx := context.Background()
	p, err := svc.Create(ctx, product.CreateInput{
		Name:        "test-product",
		DisplayName: "Test Product",
		Description: "A test product",
	})

	require.NoError(t, err)
	assert.NotEmpty(t, p.ID)
	assert.Equal(t, "test-product", p.Name)
	assert.Equal(t, "Test Product", p.DisplayName)

	// Verify main branch was auto-created
	branches, err := branchRepo.FindByProductID(ctx, p.ID, branch.ListOptions{Limit: 100})
	require.NoError(t, err)
	require.Len(t, branches, 1)
	assert.Equal(t, branch.TypeMain, branches[0].Type)
	assert.Equal(t, "main", branches[0].Name)
}

func TestCreateProduct_EmptyName(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	productRepo := product.NewSQLiteRepository(db)
	branchRepo := branch.NewSQLiteRepository(db)
	svc := product.NewService(productRepo, branchRepo)

	_, err := svc.Create(context.Background(), product.CreateInput{Name: ""})
	assert.Error(t, err)
}

func TestGetProduct(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	productRepo := product.NewSQLiteRepository(db)
	branchRepo := branch.NewSQLiteRepository(db)
	svc := product.NewService(productRepo, branchRepo)

	ctx := context.Background()
	p, _ := svc.Create(ctx, product.CreateInput{Name: "test"})

	got, err := svc.GetByID(ctx, p.ID)
	require.NoError(t, err)
	assert.Equal(t, p.ID, got.ID)
}

func TestGetProduct_NotFound(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	productRepo := product.NewSQLiteRepository(db)
	branchRepo := branch.NewSQLiteRepository(db)
	svc := product.NewService(productRepo, branchRepo)

	_, err := svc.GetByID(context.Background(), "nonexistent")
	assert.Error(t, err)
}

func TestUpdateProduct(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	productRepo := product.NewSQLiteRepository(db)
	branchRepo := branch.NewSQLiteRepository(db)
	svc := product.NewService(productRepo, branchRepo)

	ctx := context.Background()
	p, _ := svc.Create(ctx, product.CreateInput{Name: "test"})

	newName := "Updated Name"
	updated, err := svc.Update(ctx, p.ID, product.UpdateInput{DisplayName: &newName})
	require.NoError(t, err)
	assert.Equal(t, "Updated Name", updated.DisplayName)
}

func TestDeleteProduct(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	productRepo := product.NewSQLiteRepository(db)
	branchRepo := branch.NewSQLiteRepository(db)
	svc := product.NewService(productRepo, branchRepo)

	ctx := context.Background()
	p, _ := svc.Create(ctx, product.CreateInput{Name: "test"})

	err := svc.Delete(ctx, p.ID)
	require.NoError(t, err)

	_, err = svc.GetByID(ctx, p.ID)
	assert.Error(t, err)
}

func TestListProducts(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	productRepo := product.NewSQLiteRepository(db)
	branchRepo := branch.NewSQLiteRepository(db)
	svc := product.NewService(productRepo, branchRepo)

	ctx := context.Background()
	svc.Create(ctx, product.CreateInput{Name: "p1"})
	svc.Create(ctx, product.CreateInput{Name: "p2"})

	products, err := svc.List(ctx, product.DefaultListOptions())
	require.NoError(t, err)
	assert.Len(t, products, 2)
}
