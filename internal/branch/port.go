package branch

import "context"

// --- Driving Port (UseCase) ---

// UseCase defines branch business operations
type UseCase interface {
	CreateReleaseLine(ctx context.Context, productID string, input CreateReleaseLineInput) (*BranchLine, error)
	GetByID(ctx context.Context, id string) (*BranchLine, error)
	ListByProductID(ctx context.Context, productID string, opts ListOptions) ([]*BranchLine, error)
	Update(ctx context.Context, id string, input UpdateBranchInput) (*BranchLine, error)
	GetCurrentState(ctx context.Context, branchID string) (*CurrentState, error)
	UpdateCurrentState(ctx context.Context, branchID string, input UpdateCurrentStateInput) (*CurrentState, error)
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

// CurrentStateRepository defines persistence operations for CurrentState
type CurrentStateRepository interface {
	Upsert(ctx context.Context, cs *CurrentState) error
	FindByBranchID(ctx context.Context, branchID string) (*CurrentState, error)
}

// --- DTOs ---

type CreateReleaseLineInput struct {
	Name                 string  `json:"name"`
	DisplayName          string  `json:"displayName"`
	ForkedFromSnapshotID *string `json:"forkedFromSnapshotId,omitempty"`
}

type UpdateBranchInput struct {
	DisplayName *string       `json:"displayName"`
	Status      *BranchStatus `json:"status"`
}

type UpdateCurrentStateInput struct {
	RootDTProjectUUID *string `json:"rootDtProjectUuid"`
	RootBOMSerial     *string `json:"rootBomSerialNumber"`
	RootBOMVersion    *int    `json:"rootBomVersion"`
	RootBOMSHA256     *string `json:"rootBomSha256"`
	SourceRevision    *string `json:"sourceRevision"`
}

type ListOptions struct {
	Offset int
	Limit  int
}

func DefaultListOptions() ListOptions {
	return ListOptions{Offset: 0, Limit: 100}
}
