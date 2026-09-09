package version

import (
	"context"
	"time"

	"github.com/ShinyaT1994/release-catalog/internal/shared/apperror"
	"github.com/google/uuid"
)

// Service implements UseCase for versions.
type Service struct {
	repo         Repository
	branchFinder BranchFinder
}

// NewService creates a new version service.
func NewService(repo Repository, branchFinder BranchFinder) *Service {
	return &Service{repo: repo, branchFinder: branchFinder}
}

// Create presets a version. Only VersionString is required. The parent is set to
// the current latest version on the branch line. For a release branch's first
// version with ForkedFromVersionID set, the ROOT binding defaults from the forked
// main version (independently editable afterwards).
func (s *Service) Create(ctx context.Context, branchID string, input CreateInput) (*Version, error) {
	if input.VersionString == "" {
		return nil, apperror.New(apperror.CodeInvalidRequest, "versionString is required")
	}

	bi, err := s.branchFinder.FindByID(ctx, branchID)
	if err != nil {
		return nil, apperror.New(apperror.CodeInternalError, err.Error())
	}
	if bi == nil {
		return nil, apperror.New(apperror.CodeBranchNotFound, "branch not found")
	}

	latest, err := s.repo.FindLatestByBranchID(ctx, branchID)
	if err != nil {
		return nil, apperror.New(apperror.CodeInternalError, err.Error())
	}

	now := time.Now().UTC()
	v := &Version{
		ID:            uuid.New().String(),
		BranchLineID:  branchID,
		VersionString: input.VersionString,
		Status:        StatusIncomplete,
		CreatedAt:     now,
		UpdatedAt:     now,
	}
	if latest != nil {
		v.ParentVersionID = &latest.ID
	}

	// Fork handling: only meaningful for the first version on a release branch.
	isFirst := latest == nil
	if input.ForkedFromVersionID != nil && isFirst {
		forked, err := s.repo.FindByID(ctx, *input.ForkedFromVersionID)
		if err != nil {
			return nil, apperror.New(apperror.CodeInternalError, err.Error())
		}
		if forked == nil {
			return nil, apperror.New(apperror.CodeVersionNotFound, "forked-from version not found")
		}
		v.ForkedFromVersionID = input.ForkedFromVersionID
	}

	if err := s.repo.Create(ctx, v); err != nil {
		return nil, apperror.New(apperror.CodeConflict, "failed to create version: "+err.Error())
	}

	// Default ROOT from forked main version (Option 2: independently editable).
	if v.ForkedFromVersionID != nil {
		if root, err := s.repo.FindProjectByRole(ctx, *v.ForkedFromVersionID, RoleRoot); err == nil && root != nil {
			clone := &DTProject{
				VersionID:      v.ID,
				Role:           RoleRoot,
				DTProjectUUID:  root.DTProjectUUID,
				BOMSerial:      root.BOMSerial,
				BOMVersion:     root.BOMVersion,
				BOMSHA256:      root.BOMSHA256,
				SourceRevision: root.SourceRevision,
				Label:          root.Label,
				CreatedAt:      now,
				UpdatedAt:      now,
			}
			_ = s.repo.CreateProject(ctx, clone)
		}
	}

	return s.hydrate(ctx, v)
}

func (s *Service) GetByID(ctx context.Context, id string) (*Version, error) {
	v, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, apperror.New(apperror.CodeInternalError, err.Error())
	}
	if v == nil {
		return nil, apperror.New(apperror.CodeVersionNotFound, "version not found")
	}
	return s.hydrate(ctx, v)
}

func (s *Service) ListByBranchID(ctx context.Context, branchID string, opts ListOptions) ([]*Version, error) {
	bi, err := s.branchFinder.FindByID(ctx, branchID)
	if err != nil {
		return nil, apperror.New(apperror.CodeInternalError, err.Error())
	}
	if bi == nil {
		return nil, apperror.New(apperror.CodeBranchNotFound, "branch not found")
	}
	versions, err := s.repo.FindByBranchID(ctx, branchID, opts)
	if err != nil {
		return nil, apperror.New(apperror.CodeInternalError, err.Error())
	}
	for _, v := range versions {
		projects, err := s.repo.ListProjects(ctx, v.ID)
		if err != nil {
			return nil, apperror.New(apperror.CodeInternalError, err.Error())
		}
		v.Projects = projects
	}
	return versions, nil
}

// Update applies partial edits (soft immutability); every field is editable.
func (s *Service) Update(ctx context.Context, id string, input UpdateInput) (*Version, error) {
	v, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, apperror.New(apperror.CodeInternalError, err.Error())
	}
	if v == nil {
		return nil, apperror.New(apperror.CodeVersionNotFound, "version not found")
	}

	if input.VersionString != nil {
		if *input.VersionString == "" {
			return nil, apperror.New(apperror.CodeInvalidRequest, "versionString cannot be empty")
		}
		v.VersionString = *input.VersionString
	}
	if input.Status != nil {
		if *input.Status != StatusIncomplete && *input.Status != StatusFinalized {
			return nil, apperror.New(apperror.CodeInvalidRequest, "invalid status")
		}
		v.Status = *input.Status
	}
	if input.Location != nil {
		v.Location = normalizeStr(input.Location)
	}
	if input.Customer != nil {
		v.Customer = normalizeStr(input.Customer)
	}
	if input.ReleaseDate != nil {
		if *input.ReleaseDate == "" {
			v.ReleaseDate = nil
		} else {
			t, err := time.Parse(time.RFC3339, *input.ReleaseDate)
			if err != nil {
				return nil, apperror.New(apperror.CodeInvalidRequest, "releaseDate must be RFC3339")
			}
			tu := t.UTC()
			v.ReleaseDate = &tu
		}
	}
	v.UpdatedAt = time.Now().UTC()

	if err := s.repo.Update(ctx, v); err != nil {
		return nil, apperror.New(apperror.CodeConflict, err.Error())
	}
	return s.hydrate(ctx, v)
}

func (s *Service) Delete(ctx context.Context, id string) error {
	v, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return apperror.New(apperror.CodeInternalError, err.Error())
	}
	if v == nil {
		return apperror.New(apperror.CodeVersionNotFound, "version not found")
	}
	if err := s.repo.Delete(ctx, id); err != nil {
		return apperror.New(apperror.CodeInternalError, err.Error())
	}
	return nil
}

// SetRoot sets or replaces the single ROOT binding for a version.
func (s *Service) SetRoot(ctx context.Context, versionID string, input SetProjectInput) (*DTProject, error) {
	v, err := s.repo.FindByID(ctx, versionID)
	if err != nil {
		return nil, apperror.New(apperror.CodeInternalError, err.Error())
	}
	if v == nil {
		return nil, apperror.New(apperror.CodeVersionNotFound, "version not found")
	}

	now := time.Now().UTC()
	existing, err := s.repo.FindProjectByRole(ctx, versionID, RoleRoot)
	if err != nil {
		return nil, apperror.New(apperror.CodeInternalError, err.Error())
	}
	if existing != nil {
		existing.DTProjectUUID = input.DTProjectUUID
		existing.BOMSerial = input.BOMSerial
		existing.BOMVersion = input.BOMVersion
		existing.BOMSHA256 = input.BOMSHA256
		existing.SourceRevision = input.SourceRevision
		existing.Label = input.Label
		existing.UpdatedAt = now
		if err := s.repo.UpdateProject(ctx, existing); err != nil {
			return nil, apperror.New(apperror.CodeInternalError, err.Error())
		}
		return existing, nil
	}

	p := &DTProject{
		VersionID:      versionID,
		Role:           RoleRoot,
		DTProjectUUID:  input.DTProjectUUID,
		BOMSerial:      input.BOMSerial,
		BOMVersion:     input.BOMVersion,
		BOMSHA256:      input.BOMSHA256,
		SourceRevision: input.SourceRevision,
		Label:          input.Label,
		CreatedAt:      now,
		UpdatedAt:      now,
	}
	if err := s.repo.CreateProject(ctx, p); err != nil {
		return nil, apperror.New(apperror.CodeInternalError, err.Error())
	}
	return p, nil
}

// AddProject adds a PROFILE or SUB binding. ROOT must be set via SetRoot to
// enforce exactly-one-ROOT.
func (s *Service) AddProject(ctx context.Context, versionID string, input AddProjectInput) (*DTProject, error) {
	v, err := s.repo.FindByID(ctx, versionID)
	if err != nil {
		return nil, apperror.New(apperror.CodeInternalError, err.Error())
	}
	if v == nil {
		return nil, apperror.New(apperror.CodeVersionNotFound, "version not found")
	}

	switch input.Role {
	case RoleProfile, RoleSub:
		// ok
	case RoleRoot:
		return nil, apperror.New(apperror.CodeInvalidRequest, "use the ROOT endpoint to set the root project")
	default:
		return nil, apperror.New(apperror.CodeInvalidRequest, "role must be PROFILE or SUB")
	}

	now := time.Now().UTC()
	p := &DTProject{
		VersionID:      versionID,
		Role:           input.Role,
		DTProjectUUID:  input.DTProjectUUID,
		BOMSerial:      input.BOMSerial,
		BOMVersion:     input.BOMVersion,
		BOMSHA256:      input.BOMSHA256,
		SourceRevision: input.SourceRevision,
		Label:          input.Label,
		CreatedAt:      now,
		UpdatedAt:      now,
	}
	if err := s.repo.CreateProject(ctx, p); err != nil {
		return nil, apperror.New(apperror.CodeInternalError, err.Error())
	}
	return p, nil
}

func (s *Service) RemoveProject(ctx context.Context, versionID string, bindingID int64) error {
	p, err := s.repo.FindProject(ctx, versionID, bindingID)
	if err != nil {
		return apperror.New(apperror.CodeInternalError, err.Error())
	}
	if p == nil {
		return apperror.New(apperror.CodeVersionNotFound, "binding not found")
	}
	if err := s.repo.DeleteProject(ctx, versionID, bindingID); err != nil {
		return apperror.New(apperror.CodeInternalError, err.Error())
	}
	return nil
}

func (s *Service) hydrate(ctx context.Context, v *Version) (*Version, error) {
	projects, err := s.repo.ListProjects(ctx, v.ID)
	if err != nil {
		return nil, apperror.New(apperror.CodeInternalError, err.Error())
	}
	v.Projects = projects
	return v, nil
}

func normalizeStr(s *string) *string {
	if s == nil || *s == "" {
		return nil
	}
	return s
}
