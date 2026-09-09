package main

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/ShinyaT1994/release-catalog/internal/branch"
	"github.com/ShinyaT1994/release-catalog/internal/dtclient"
	"github.com/ShinyaT1994/release-catalog/internal/dtproxy"
	"github.com/ShinyaT1994/release-catalog/internal/graph"
	"github.com/ShinyaT1994/release-catalog/internal/product"
	"github.com/ShinyaT1994/release-catalog/internal/shared/config"
	"github.com/ShinyaT1994/release-catalog/internal/shared/database"
	"github.com/ShinyaT1994/release-catalog/internal/shared/middleware"
	"github.com/ShinyaT1994/release-catalog/internal/version"
	"github.com/labstack/echo/v4"
	_ "github.com/mattn/go-sqlite3"
)

func main() {
	cfg := config.Load()

	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: cfg.LogLevel()}))
	slog.SetDefault(logger)

	db, err := sql.Open("sqlite3", cfg.DatabaseDSN+"?_journal_mode=WAL&_foreign_keys=on")
	if err != nil {
		slog.Error("failed to open database", "error", err)
		os.Exit(1)
	}
	defer db.Close()

	if err := database.Migrate(db); err != nil {
		slog.Error("failed to run migrations", "error", err)
		os.Exit(1)
	}

	// --- DT Client ---
	var dt dtclient.Client
	if cfg.DTStubMode {
		dt = dtclient.NewStubClient()
		slog.Info("using DT stub client")
	} else {
		dt = dtclient.NewHTTPClient(cfg.DTBaseURL, cfg.DTAPIKey, cfg.DTTimeout)
		slog.Info("using DT HTTP client", "baseURL", cfg.DTBaseURL)
	}

	// --- Repositories (Adapters: out) ---
	productRepo := product.NewSQLiteRepository(db)
	branchRepo := branch.NewSQLiteRepository(db)
	versionRepo := version.NewSQLiteRepository(db)
	lineageSource := graph.NewSQLiteLineageSource(db)

	// --- Cross-feature adapters ---
	productFinderAdapter := &productFinderAdapter{repo: productRepo}
	versionFinderAdapter := &versionFinderAdapter{repo: versionRepo}
	branchFinderAdapter := &branchFinderAdapter{repo: branchRepo}
	versionProjectFinderAdapter := &versionProjectFinderAdapter{repo: versionRepo}

	// --- UseCases ---
	productUC := product.NewService(productRepo, branchRepo)
	versionUC := version.NewService(versionRepo, branchFinderAdapter)
	versionCreatorAdapter := &versionCreatorAdapter{uc: versionUC}
	branchUC := branch.NewService(branchRepo, productFinderAdapter, versionFinderAdapter, versionCreatorAdapter)
	sbomUC := graph.NewSBOMService(dt, versionProjectFinderAdapter)
	lineageUC := graph.NewLineageService(lineageSource)

	// --- Echo setup ---
	e := echo.New()
	e.HideBanner = true
	e.Use(middleware.RequestID())
	e.Use(middleware.Logger())
	e.Use(middleware.CORS(cfg.CORSAllowOrigins))
	e.Use(middleware.AuthPlaceholder())

	api := e.Group("/api/v1")

	// --- Register routes (package-by-feature) ---
	product.NewHandler(productUC).RegisterRoutes(api)
	branch.NewHandler(branchUC).RegisterRoutes(api)
	version.NewHandler(versionUC).RegisterRoutes(api)
	graph.NewHandler(sbomUC, lineageUC).RegisterRoutes(api)
	dtproxy.NewHandler(dt).RegisterRoutes(api)

	// Health
	e.GET("/health", func(c echo.Context) error {
		return c.JSON(200, map[string]string{"status": "ok"})
	})

	// --- Start ---
	addr := fmt.Sprintf(":%d", cfg.ServerPort)
	slog.Info("starting server", "addr", addr)

	go func() {
		if err := e.Start(addr); err != nil && err != http.ErrServerClosed {
			slog.Error("server error", "error", err)
			os.Exit(1)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt)
	<-quit

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := e.Shutdown(ctx); err != nil {
		slog.Error("shutdown error", "error", err)
	}
	slog.Info("server stopped")
}

// --- Cross-feature adapter implementations ---

type productFinderAdapter struct {
	repo *product.SQLiteRepository
}

func (a *productFinderAdapter) FindByID(ctx context.Context, id string) (bool, error) {
	p, err := a.repo.FindByID(ctx, id)
	if err != nil {
		return false, err
	}
	return p != nil, nil
}

// versionFinderAdapter lets the branch feature validate a fork-point version.
type versionFinderAdapter struct {
	repo *version.SQLiteRepository
}

func (a *versionFinderAdapter) FindByID(ctx context.Context, id string) (*branch.VersionInfo, error) {
	v, err := a.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if v == nil {
		return nil, nil
	}
	return &branch.VersionInfo{ID: v.ID, BranchLineID: v.BranchLineID, VersionString: v.VersionString}, nil
}

// branchFinderAdapter lets the version feature verify branch existence/type.
type branchFinderAdapter struct {
	repo *branch.SQLiteRepository
}

func (a *branchFinderAdapter) FindByID(ctx context.Context, id string) (*version.BranchInfo, error) {
	b, err := a.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if b == nil {
		return nil, nil
	}
	return &version.BranchInfo{ID: b.ID, Type: string(b.Type)}, nil
}

// versionProjectFinderAdapter lets the SBOM graph resolve a role-tagged DT project.
type versionProjectFinderAdapter struct {
	repo *version.SQLiteRepository
}

func (a *versionProjectFinderAdapter) VersionExists(ctx context.Context, versionID string) (bool, error) {
	v, err := a.repo.FindByID(ctx, versionID)
	if err != nil {
		return false, err
	}
	return v != nil, nil
}

func (a *versionProjectFinderAdapter) FindProjectUUID(ctx context.Context, versionID, role string) (*string, error) {
	p, err := a.repo.FindProjectByRole(ctx, versionID, version.Role(role))
	if err != nil {
		return nil, err
	}
	if p == nil {
		return nil, nil
	}
	return p.DTProjectUUID, nil
}

// versionCreatorAdapter lets the branch feature auto-create the first version on
// a newly created release branch, delegating to the version usecase (which
// defaults the ROOT binding from the forked main version, status "incomplete").
type versionCreatorAdapter struct {
	uc version.UseCase
}

func (a *versionCreatorAdapter) CreateFirstVersion(ctx context.Context, branchID string, input branch.CreateFirstVersionInput) error {
	_, err := a.uc.Create(ctx, branchID, version.CreateInput{
		VersionString:       input.VersionString,
		ForkedFromVersionID: input.ForkedFromVersionID,
	})
	return err
}
