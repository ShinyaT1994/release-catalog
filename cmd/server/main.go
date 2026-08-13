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
	"github.com/ShinyaT1994/release-catalog/internal/graph"
	"github.com/ShinyaT1994/release-catalog/internal/product"
	"github.com/ShinyaT1994/release-catalog/internal/release"
	"github.com/ShinyaT1994/release-catalog/internal/shared/config"
	"github.com/ShinyaT1994/release-catalog/internal/shared/database"
	"github.com/ShinyaT1994/release-catalog/internal/shared/middleware"
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
	csRepo := branch.NewSQLiteCurrentStateRepository(db)
	releaseRepo := release.NewSQLiteRepository(db)

	// --- Adapters for cross-feature dependencies ---
	productFinderAdapter := &productFinderAdapter{repo: productRepo}
	snapshotFinderAdapter := &snapshotFinderAdapter{repo: releaseRepo}
	branchFinderAdapter := &branchFinderAdapter{repo: branchRepo}
	csFinderAdapter := &csFinderAdapter{repo: csRepo}
	graphCSFinderAdapter := &graphCSFinderAdapter{repo: csRepo, branchRepo: branchRepo}
	graphSnapFinderAdapter := &graphSnapFinderAdapter{repo: releaseRepo}

	// --- UseCases ---
	productUC := product.NewService(productRepo, branchRepo)
	branchUC := branch.NewService(branchRepo, csRepo, productFinderAdapter, snapshotFinderAdapter)
	releaseUC := release.NewService(releaseRepo, branchFinderAdapter, csFinderAdapter)
	graphUC := graph.NewService(dt, graphCSFinderAdapter, graphSnapFinderAdapter)

	// --- Echo setup ---
	e := echo.New()
	e.HideBanner = true
	e.Use(middleware.RequestID())
	e.Use(middleware.Logger())
	e.Use(middleware.AuthPlaceholder())

	api := e.Group("/api/v1")

	// --- Register routes (package-by-feature) ---
	product.NewHandler(productUC).RegisterRoutes(api)
	branch.NewHandler(branchUC).RegisterRoutes(api)
	release.NewHandler(releaseUC).RegisterRoutes(api)
	graph.NewHandler(graphUC).RegisterRoutes(api)

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

type snapshotFinderAdapter struct {
	repo *release.SQLiteRepository
}

func (a *snapshotFinderAdapter) FindByID(ctx context.Context, id string) (*branch.SnapshotInfo, error) {
	s, err := a.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if s == nil {
		return nil, nil
	}
	return &branch.SnapshotInfo{
		ID:                s.ID,
		RootDTProjectUUID: s.RootDTProjectUUID,
		RootBOMSerial:     s.RootBOMSerial,
		RootBOMVersion:    s.RootBOMVersion,
		RootBOMSHA256:     s.RootBOMSHA256,
		SourceRevision:    s.SourceRevision,
	}, nil
}

type branchFinderAdapter struct {
	repo *branch.SQLiteRepository
}

func (a *branchFinderAdapter) FindByID(ctx context.Context, id string) (*release.BranchInfo, error) {
	b, err := a.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if b == nil {
		return nil, nil
	}
	return &release.BranchInfo{ID: b.ID, Type: string(b.Type)}, nil
}

type csFinderAdapter struct {
	repo *branch.SQLiteCurrentStateRepository
}

func (a *csFinderAdapter) FindByBranchID(ctx context.Context, branchID string) (*release.CurrentStateInfo, error) {
	cs, err := a.repo.FindByBranchID(ctx, branchID)
	if err != nil {
		return nil, err
	}
	if cs == nil {
		return nil, nil
	}
	return &release.CurrentStateInfo{
		RootDTProjectUUID: cs.RootDTProjectUUID,
		RootBOMSerial:     cs.RootBOMSerial,
		RootBOMVersion:    cs.RootBOMVersion,
		RootBOMSHA256:     cs.RootBOMSHA256,
		SourceRevision:    cs.SourceRevision,
	}, nil
}

type graphCSFinderAdapter struct {
	repo       *branch.SQLiteCurrentStateRepository
	branchRepo *branch.SQLiteRepository
}

func (a *graphCSFinderAdapter) FindByBranchID(ctx context.Context, branchID string) (*graph.CurrentStateInfo, error) {
	cs, err := a.repo.FindByBranchID(ctx, branchID)
	if err != nil {
		return nil, err
	}
	if cs == nil {
		return nil, nil
	}
	return &graph.CurrentStateInfo{RootDTProjectUUID: cs.RootDTProjectUUID}, nil
}

func (a *graphCSFinderAdapter) BranchExists(ctx context.Context, branchID string) (bool, error) {
	b, err := a.branchRepo.FindByID(ctx, branchID)
	if err != nil {
		return false, err
	}
	return b != nil, nil
}

type graphSnapFinderAdapter struct {
	repo *release.SQLiteRepository
}

func (a *graphSnapFinderAdapter) FindByID(ctx context.Context, id string) (*graph.SnapshotInfo, error) {
	s, err := a.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if s == nil {
		return nil, nil
	}
	return &graph.SnapshotInfo{RootDTProjectUUID: s.RootDTProjectUUID}, nil
}
