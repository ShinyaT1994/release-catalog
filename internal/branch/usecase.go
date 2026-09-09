package branch

import (
	"context"
	"time"

	"github.com/ShinyaT1994/release-catalog/internal/shared/apperror"
	"github.com/google/uuid"
)

// Service implements UseCase for Branch
type Service struct {
	repo           Repository
	productFinder  ProductFinder
	versionFinder  VersionFinder
	versionCreator VersionCreator
}

// NewService creates a new branch service. versionCreator may be nil (in which
// case no first version is auto-created on release-line creation).
func NewService(repo Repository, productFinder ProductFinder, versionFinder VersionFinder, versionCreator VersionCreator) *Service {
	return &Service{repo: repo, productFinder: productFinder, versionFinder: versionFinder, versionCreator: versionCreator}
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

	// Validate the fork-point version if provided: it must exist and belong to
	// this product's main branch.
	var forkedVersionString string
	if input.ForkedFromVersionID != nil {
		vi, err := s.versionFinder.FindByID(ctx, *input.ForkedFromVersionID)
		if err != nil {
			return nil, apperror.New(apperror.CodeInternalError, err.Error())
		}
		if vi == nil {
			return nil, apperror.New(apperror.CodeVersionNotFound, "forked-from version not found")
		}
		if vi.BranchLineID != mainBranch.ID {
			return nil, apperror.New(apperror.CodeInvalidRequest, "forked-from version must belong to the main branch")
		}
		forkedVersionString = vi.VersionString
	}

	now := time.Now().UTC()
	b := &BranchLine{
		ID:                  uuid.New().String(),
		ProductID:           productID,
		Type:                TypeRelease,
		Name:                input.Name,
		DisplayName:         input.DisplayName,
		SourceBranchLineID:  &mainBranch.ID,
		ForkedFromVersionID: input.ForkedFromVersionID,
		Status:              StatusActive,
		CreatedAt:           now,
		UpdatedAt:           now,
	}

	if err := s.repo.Create(ctx, b); err != nil {
		return nil, apperror.New(apperror.CodeConflict, "failed to create release line: "+err.Error())
	}

	// Auto-create the release branch's first version when it forks from a main
	// version. The version usecase defaults the ROOT binding from the forked
	// main version and sets status "incomplete"; users then add PROFILE/SUB
	// bindings. The initial versionString inherits the forked main version's
	// string.
	if input.ForkedFromVersionID != nil && s.versionCreator != nil {
		vs := forkedVersionString
		if vs == "" {
			vs = "0.1.0"
		}
		if err := s.versionCreator.CreateFirstVersion(ctx, b.ID, CreateFirstVersionInput{
			VersionString:       vs,
			ForkedFromVersionID: input.ForkedFromVersionID,
		}); err != nil {
			return nil, apperror.New(apperror.CodeInternalError, "failed to create initial version: "+err.Error())
		}
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
