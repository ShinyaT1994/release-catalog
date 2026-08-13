package release

import (
	"context"
	"time"

	"github.com/ShinyaT1994/release-catalog/internal/shared/apperror"
	"github.com/google/uuid"
)

// Service implements UseCase for Release
type Service struct {
	repo     Repository
	branchFinder BranchFinder
	csFinder     CurrentStateFinder
}

// NewService creates a new release service
func NewService(repo Repository, branchFinder BranchFinder, csFinder CurrentStateFinder) *Service {
	return &Service{repo: repo, branchFinder: branchFinder, csFinder: csFinder}
}

func (s *Service) CreateMainSnapshot(ctx context.Context, branchID string, input CreateSnapshotInput) (*Snapshot, error) {
	if input.Version == "" {
		return nil, apperror.New(apperror.CodeInvalidRequest, "version is required")
	}

	bi, err := s.branchFinder.FindByID(ctx, branchID)
	if err != nil {
		return nil, apperror.New(apperror.CodeInternalError, err.Error())
	}
	if bi == nil {
		return nil, apperror.New(apperror.CodeBranchNotFound, "branch not found")
	}
	if bi.Type != "MAIN" {
		return nil, apperror.New(apperror.CodeInvalidRequest, "main snapshots can only be created on main branches")
	}

	cs, _ := s.csFinder.FindByBranchID(ctx, branchID)

	now := time.Now().UTC()
	snap := &Snapshot{
		ID:           uuid.New().String(),
		BranchLineID: branchID,
		SnapshotType: TypeMainSnapshot,
		Version:      input.Version,
		Status:       StatusReleased,
		CreatedAt:    now,
	}

	if cs != nil {
		snap.RootDTProjectUUID = cs.RootDTProjectUUID
		snap.RootBOMSerial = cs.RootBOMSerial
		snap.RootBOMVersion = cs.RootBOMVersion
		snap.RootBOMSHA256 = cs.RootBOMSHA256
		snap.SourceRevision = cs.SourceRevision
	}

	if err := s.repo.Create(ctx, snap); err != nil {
		return nil, apperror.New(apperror.CodeInternalError, "failed to create snapshot: "+err.Error())
	}
	return snap, nil
}

func (s *Service) CreateRelease(ctx context.Context, branchID string, input CreateReleaseInput) (*Snapshot, error) {
	if input.Version == "" {
		return nil, apperror.New(apperror.CodeInvalidRequest, "version is required")
	}

	bi, err := s.branchFinder.FindByID(ctx, branchID)
	if err != nil {
		return nil, apperror.New(apperror.CodeInternalError, err.Error())
	}
	if bi == nil {
		return nil, apperror.New(apperror.CodeBranchNotFound, "branch not found")
	}
	if bi.Type != "RELEASE" {
		return nil, apperror.New(apperror.CodeInvalidRequest, "releases can only be created on release branches")
	}

	cs, _ := s.csFinder.FindByBranchID(ctx, branchID)

	now := time.Now().UTC()
	snap := &Snapshot{
		ID:           uuid.New().String(),
		BranchLineID: branchID,
		SnapshotType: TypeRelease,
		Version:      input.Version,
		Status:       StatusReleased,
		ReleasedAt:   &now,
		CreatedAt:    now,
	}

	if cs != nil {
		snap.RootDTProjectUUID = cs.RootDTProjectUUID
		snap.RootBOMSerial = cs.RootBOMSerial
		snap.RootBOMVersion = cs.RootBOMVersion
		snap.RootBOMSHA256 = cs.RootBOMSHA256
		snap.SourceRevision = cs.SourceRevision
	}

	if err := s.repo.Create(ctx, snap); err != nil {
		return nil, apperror.New(apperror.CodeInternalError, "failed to create release: "+err.Error())
	}
	return snap, nil
}

func (s *Service) GetByID(ctx context.Context, id string) (*Snapshot, error) {
	snap, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, apperror.New(apperror.CodeInternalError, err.Error())
	}
	if snap == nil {
		return nil, apperror.New(apperror.CodeReleaseNotFound, "release not found")
	}
	return snap, nil
}

func (s *Service) ListByBranchID(ctx context.Context, branchID string, opts ListOptions) ([]*Snapshot, error) {
	bi, err := s.branchFinder.FindByID(ctx, branchID)
	if err != nil {
		return nil, apperror.New(apperror.CodeInternalError, err.Error())
	}
	if bi == nil {
		return nil, apperror.New(apperror.CodeBranchNotFound, "branch not found")
	}
	return s.repo.FindByBranchID(ctx, branchID, opts)
}
