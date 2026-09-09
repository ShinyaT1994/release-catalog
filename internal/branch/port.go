package branch

import "context"

// --- Driving Port (UseCase) ---

// UseCase defines branch business operations
type UseCase interface {
	CreateReleaseLine(ctx context.Context, productID string, input CreateReleaseLineInput) (*BranchLine, error)
	GetByID(ctx context.Context, id string) (*BranchLine, error)
	ListByProductID(ctx context.Context, productID string, opts ListOptions) ([]*BranchLine, error)
	Update(ctx context.Context, id string, input UpdateBranchInput) (*BranchLine, error)
}

// --- Driven Port (Repository) ---

// Repository defines persistence operations for BranchLine
type Repository interface {
	Create(ctx context.Context, b *BranchLine) error
	FindByID(ctx context.Context, id string) (*BranchLine, error)
	FindByProductID(ctx context.Context, productID string, opts ListOptions) ([]*BranchLine, error)
	FindMainByProductID(ctx context.Context, productID string) (*BranchLine, error)
	Update(ctx context.Context, b *BranchLine) error
}

// ProductFinder is used to verify product existence (avoids circular import)
type ProductFinder interface {
	FindByID(ctx context.Context, id string) (bool, error)
}

// VersionFinder verifies that a fork-point version exists and belongs to the
// product's main branch (avoids circular import).
type VersionFinder interface {
	FindByID(ctx context.Context, id string) (*VersionInfo, error)
}

// VersionInfo is the minimal version info needed for fork validation.
type VersionInfo struct {
	ID            string
	BranchLineID  string
	VersionString string
}

// VersionCreator auto-creates the first version on a newly created release
// branch (avoids circular import). Implemented by an adapter delegating to the
// version usecase's Create, which defaults the ROOT binding from the forked
// main version and sets status "incomplete".
type VersionCreator interface {
	CreateFirstVersion(ctx context.Context, branchID string, input CreateFirstVersionInput) error
}

// CreateFirstVersionInput carries the data needed to preset the first version
// on a release branch.
type CreateFirstVersionInput struct {
	VersionString       string
	ForkedFromVersionID *string
}

// --- DTOs ---

type CreateReleaseLineInput struct {
	Name                string  `json:"name"`
	DisplayName         string  `json:"displayName"`
	ForkedFromVersionID *string `json:"forkedFromVersionId,omitempty"`
}

type UpdateBranchInput struct {
	DisplayName *string       `json:"displayName"`
	Status      *BranchStatus `json:"status"`
}

type ListOptions struct {
	Offset int
	Limit  int
}

func DefaultListOptions() ListOptions {
	return ListOptions{Offset: 0, Limit: 100}
}
