package release

import "context"

// --- Driving Port (UseCase) ---

// UseCase defines release/snapshot business operations
type UseCase interface {
	CreateMainSnapshot(ctx context.Context, branchID string, input CreateSnapshotInput) (*Snapshot, error)
	CreateRelease(ctx context.Context, branchID string, input CreateReleaseInput) (*Snapshot, error)
	GetByID(ctx context.Context, id string) (*Snapshot, error)
	ListByBranchID(ctx context.Context, branchID string, opts ListOptions) ([]*Snapshot, error)
}

// --- Driven Port (Repository) ---

// Repository defines persistence operations for Snapshot
type Repository interface {
	Create(ctx context.Context, s *Snapshot) error
	FindByID(ctx context.Context, id string) (*Snapshot, error)
	FindByBranchID(ctx context.Context, branchID string, opts ListOptions) ([]*Snapshot, error)
}

// BranchFinder is used to verify branch existence (avoids circular import)
type BranchFinder interface {
	FindByID(ctx context.Context, id string) (*BranchInfo, error)
}

// CurrentStateFinder is used to get current state for snapshotting
type CurrentStateFinder interface {
	FindByBranchID(ctx context.Context, branchID string) (*CurrentStateInfo, error)
}

// BranchInfo minimal branch info needed
type BranchInfo struct {
	ID   string
	Type string // "MAIN" or "RELEASE"
}

// CurrentStateInfo minimal current state info
type CurrentStateInfo struct {
	RootDTProjectUUID *string
	RootBOMSerial     *string
	RootBOMVersion    *int
	RootBOMSHA256     *string
	SourceRevision    *string
}

// --- DTOs ---

type CreateSnapshotInput struct {
	Version string `json:"version"`
}

type CreateReleaseInput struct {
	Version string `json:"version"`
}

type ListOptions struct {
	Offset int
	Limit  int
}

func DefaultListOptions() ListOptions {
	return ListOptions{Offset: 0, Limit: 100}
}
