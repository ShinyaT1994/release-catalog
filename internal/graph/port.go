package graph

import "context"

// --- SBOM graph (BOM-Link resolution via Dependency-Track) ---

// SBOMUseCase resolves the BOM-Link SBOM graph for a role-tagged project.
type SBOMUseCase interface {
	GetProjectGraph(ctx context.Context, versionID, role string, opts Options) (*ReleaseGraph, error)
}

// VersionProjectFinder resolves a role-tagged project's DT UUID for a version
// (avoids circular import with the version feature).
type VersionProjectFinder interface {
	VersionExists(ctx context.Context, versionID string) (bool, error)
	FindProjectUUID(ctx context.Context, versionID, role string) (*string, error)
}

// --- Lineage graph + timeline (RC DB only) ---

// LineageUseCase builds the version-lineage graph and timelines from RC DB.
type LineageUseCase interface {
	GetProductLineage(ctx context.Context, productID string, axis LineageAxis) (*LineageGraph, error)
	GetProductTimeline(ctx context.Context, productID string) (*Timeline, error)
	GetBranchTimeline(ctx context.Context, branchID string) (*Timeline, error)
}

// LineageSource provides read-only access to versions and branches for lineage.
type LineageSource interface {
	ProductExists(ctx context.Context, productID string) (bool, error)
	BranchExists(ctx context.Context, branchID string) (bool, error)
	ListVersionsByProduct(ctx context.Context, productID string) ([]VersionRow, error)
	ListVersionsByBranch(ctx context.Context, branchID string) ([]VersionRow, error)
}

// VersionRow is a flattened version + branch record for lineage building.
type VersionRow struct {
	VersionID           string
	BranchLineID        string
	BranchName          string
	BranchType          string
	VersionString       string
	Status              string
	ParentVersionID     *string
	ForkedFromVersionID *string
	ReleaseDate         *string
	CreatedAt           string
}
