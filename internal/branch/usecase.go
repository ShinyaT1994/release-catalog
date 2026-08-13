package branch

import (
	"context"
	"time"

	"github.com/ShinyaT1994/release-catalog/internal/shared/apperror"
	"github.com/google/uuid"
)

// Service implements UseCase for Branch
type Service struct {
	repo       Repository
	csRepo     CurrentStateRepository
	productFinder ProductFinder
	snapshotFinder SnapshotFinder
}

// ProductFinder is used to verify product existence (avoids circular import)
type ProductFinder interface {
	FindByID(ctx context.Context, id string) (bool, error)
}

// SnapshotFinder is used to find a snapshot for fork point (avoids circular import)
type SnapshotFinder interface {
	FindByID(ctx context.Context, id string) (*SnapshotInfo, error)
}

// SnapshotInfo contains the minimal info needed from a snapshot
type SnapshotInfo struct {
	ID                string
	RootDTProjectUUID *string
	RootBOMSerial     *string
	RootBOMVersion    *int
	RootBOMSHA256     *string
	SourceRevision    *string
}

// NewService creates a new branch service
func NewService(repo Repository, csRepo CurrentStateRepository, productFinder ProductFinder, snapshotFinder SnapshotFinder) *Service {
	return &Service{repo: repo, csRepo: csRepo, productFinder: productFinder, snapshotFinder: snapshotFinder}
}

func (s *Service) CreateReleaseLine(ctx context.Context, productID string, input CreateReleaseLineInput) (*BranchLine, error) {
	if input.Name == "" {
		return nil, apperror.New(apperror.CodeInvalidRequest, "name is required")
	}

	exists, err := s.productFinder.FindByID(ctx, productID)
	if err != nil {
		return nil, apperror.New(apperror.CodeInternalError, err.Error())
	}
	if !exists {
		return nil, apperror.New(apperror.CodeProductNotFound, "product not found")
	}

	mainBranch, err := s.repo.FindMainByProductID(ctx, productID)
	if err != nil {
		return nil, apperror.New(apperror.CodeInternalError, err.Error())
	}
	if mainBranch == nil {
		return nil, apperror.New(apperror.CodeInternalError, "main branch not found")
	}

	now := time.Now().UTC()
	b := &BranchLine{
		ID:                 uuid.New().String(),
		ProductID:          productID,
		Type:               TypeRelease,
		Name:               input.Name,
		DisplayName:        input.DisplayName,
		SourceBranchLineID: &mainBranch.ID,
		ForkedFromSnapshotID: input.ForkedFromSnapshotID,
		Status:             StatusActive,
		CreatedAt:          now,
		UpdatedAt:          now,
	}

	if err := s.repo.Create(ctx, b); err != nil {
		return nil, apperror.New(apperror.CodeInternalError, "failed to create release line: "+err.Error())
	}

	// Initialize current state - copy from fork source
	cs := &CurrentState{BranchLineID: b.ID, UpdatedAt: now}

	if input.ForkedFromSnapshotID != nil {
		snap, err := s.snapshotFinder.FindByID(ctx, *input.ForkedFromSnapshotID)
		if err == nil && snap != nil {
			cs.RootDTProjectUUID = snap.RootDTProjectUUID
			cs.RootBOMSerial = snap.RootBOMSerial
			cs.RootBOMVersion = snap.RootBOMVersion
			cs.RootBOMSHA256 = snap.RootBOMSHA256
			cs.SourceRevision = snap.SourceRevision
		}
	} else {
		mainCS, _ := s.csRepo.FindByBranchID(ctx, mainBranch.ID)
		if mainCS != nil {
			cs.RootDTProjectUUID = mainCS.RootDTProjectUUID
			cs.RootBOMSerial = mainCS.RootBOMSerial
			cs.RootBOMVersion = mainCS.RootBOMVersion
			cs.RootBOMSHA256 = mainCS.RootBOMSHA256
			cs.SourceRevision = mainCS.SourceRevision
		}
	}

	if err := s.csRepo.Upsert(ctx, cs); err != nil {
		return nil, apperror.New(apperror.CodeInternalError, "failed to initialize current state: "+err.Error())
	}

	return b, nil
}

func (s *Service) GetByID(ctx context.Context, id string) (*BranchLine, error) {
	b, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, apperror.New(apperror.CodeInternalError, err.Error())
	}
	if b == nil {
		return nil, apperror.New(apperror.CodeBranchNotFound, "branch not found")
	}
	return b, nil
}

func (s *Service) ListByProductID(ctx context.Context, productID string, opts ListOptions) ([]*BranchLine, error) {
	exists, err := s.productFinder.FindByID(ctx, productID)
	if err != nil {
		return nil, apperror.New(apperror.CodeInternalError, err.Error())
	}
	if !exists {
		return nil, apperror.New(apperror.CodeProductNotFound, "product not found")
	}
	return s.repo.FindByProductID(ctx, productID, opts)
}

func (s *Service) Update(ctx context.Context, id string, input UpdateBranchInput) (*BranchLine, error) {
	b, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, apperror.New(apperror.CodeInternalError, err.Error())
	}
	if b == nil {
		return nil, apperror.New(apperror.CodeBranchNotFound, "branch not found")
	}

	if input.DisplayName != nil {
		b.DisplayName = *input.DisplayName
	}
	if input.Status != nil {
		b.Status = *input.Status
		if *input.Status == StatusClosed {
			now := time.Now().UTC()
			b.ClosedAt = &now
		}
	}
	b.UpdatedAt = time.Now().UTC()

	if err := s.repo.Update(ctx, b); err != nil {
		return nil, apperror.New(apperror.CodeInternalError, err.Error())
	}
	return b, nil
}

func (s *Service) GetCurrentState(ctx context.Context, branchID string) (*CurrentState, error) {
	b, err := s.repo.FindByID(ctx, branchID)
	if err != nil {
		return nil, apperror.New(apperror.CodeInternalError, err.Error())
	}
	if b == nil {
		return nil, apperror.New(apperror.CodeBranchNotFound, "branch not found")
	}

	cs, err := s.csRepo.FindByBranchID(ctx, branchID)
	if err != nil {
		return nil, apperror.New(apperror.CodeInternalError, err.Error())
	}
	if cs == nil {
		return &CurrentState{BranchLineID: branchID, UpdatedAt: time.Now().UTC()}, nil
	}
	return cs, nil
}

func (s *Service) UpdateCurrentState(ctx context.Context, branchID string, input UpdateCurrentStateInput) (*CurrentState, error) {
	b, err := s.repo.FindByID(ctx, branchID)
	if err != nil {
		return nil, apperror.New(apperror.CodeInternalError, err.Error())
	}
	if b == nil {
		return nil, apperror.New(apperror.CodeBranchNotFound, "branch not found")
	}

	cs := &CurrentState{
		BranchLineID:      branchID,
		RootDTProjectUUID: input.RootDTProjectUUID,
		RootBOMSerial:     input.RootBOMSerial,
		RootBOMVersion:    input.RootBOMVersion,
		RootBOMSHA256:     input.RootBOMSHA256,
		SourceRevision:    input.SourceRevision,
		UpdatedAt:         time.Now().UTC(),
	}

	if err := s.csRepo.Upsert(ctx, cs); err != nil {
		return nil, apperror.New(apperror.CodeInternalError, err.Error())
	}
	return cs, nil
}
