package graph

import "context"

// UseCase defines graph business operations
type UseCase interface {
	GetBranchCurrentGraph(ctx context.Context, branchID string, opts Options) (*ReleaseGraph, error)
	GetReleaseGraph(ctx context.Context, releaseID string, opts Options) (*ReleaseGraph, error)
}

// BranchCurrentStateFinder finds branch current state (avoids circular import)
type BranchCurrentStateFinder interface {
	FindByBranchID(ctx context.Context, branchID string) (*CurrentStateInfo, error)
	BranchExists(ctx context.Context, branchID string) (bool, error)
}

// SnapshotFinder finds snapshot info
type SnapshotFinder interface {
	FindByID(ctx context.Context, id string) (*SnapshotInfo, error)
}

// CurrentStateInfo minimal info for graph resolution
type CurrentStateInfo struct {
	RootDTProjectUUID *string
}

// SnapshotInfo minimal info for graph resolution
type SnapshotInfo struct {
	RootDTProjectUUID *string
}
