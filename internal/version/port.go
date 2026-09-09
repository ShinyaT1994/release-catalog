package version

import "context"

// --- Driving Port (UseCase) ---

// UseCase defines version business operations.
type UseCase interface {
	Create(ctx context.Context, branchID string, input CreateInput) (*Version, error)
	GetByID(ctx context.Context, id string) (*Version, error)
	ListByBranchID(ctx context.Context, branchID string, opts ListOptions) ([]*Version, error)
	Update(ctx context.Context, id string, input UpdateInput) (*Version, error)
	Delete(ctx context.Context, id string) error

	// Role-tagged DT projects.
	SetRoot(ctx context.Context, versionID string, input SetProjectInput) (*DTProject, error)
	AddProject(ctx context.Context, versionID string, input AddProjectInput) (*DTProject, error)
	RemoveProject(ctx context.Context, versionID string, bindingID int64) error
}

// --- Driven Port (Repository) ---

// Repository defines persistence operations for versions and their bindings.
type Repository interface {
	Create(ctx context.Context, v *Version) error
	FindByID(ctx context.Context, id string) (*Version, error)
	FindByBranchID(ctx context.Context, branchID string, opts ListOptions) ([]*Version, error)
	Update(ctx context.Context, v *Version) error
	Delete(ctx context.Context, id string) error

	// Latest version on a branch line (by release datetime, then created_at).
	FindLatestByBranchID(ctx context.Context, branchID string) (*Version, error)

	// Role-tagged bindings.
	ListProjects(ctx context.Context, versionID string) ([]*DTProject, error)
	FindProject(ctx context.Context, versionID string, bindingID int64) (*DTProject, error)
	FindProjectByRole(ctx context.Context, versionID string, role Role) (*DTProject, error)
	CreateProject(ctx context.Context, p *DTProject) error
	UpdateProject(ctx context.Context, p *DTProject) error
	DeleteProject(ctx context.Context, versionID string, bindingID int64) error
}

// BranchFinder verifies branch existence and type (avoids circular import).
type BranchFinder interface {
	FindByID(ctx context.Context, id string) (*BranchInfo, error)
}

// BranchInfo is the minimal branch info needed by the version usecase.
type BranchInfo struct {
	ID   string
	Type string // "MAIN" or "RELEASE"
}

// --- DTOs ---

// CreateInput is a partial ("preset") create; only VersionString is required.
type CreateInput struct {
	VersionString       string  `json:"versionString"`
	ForkedFromVersionID *string `json:"forkedFromVersionId,omitempty"`
}

// UpdateInput edits any field (soft immutability). Nil pointers are unchanged.
type UpdateInput struct {
	VersionString *string `json:"versionString"`
	Status        *Status `json:"status"`
	Location      *string `json:"location"`
	Customer      *string `json:"customer"`
	ReleaseDate   *string `json:"releaseDate"` // RFC3339, empty string clears
}

// SetProjectInput sets/replaces the single ROOT binding.
type SetProjectInput struct {
	DTProjectUUID  *string `json:"dtProjectUuid"`
	BOMSerial      *string `json:"bomSerialNumber"`
	BOMVersion     *int    `json:"bomVersion"`
	BOMSHA256      *string `json:"bomSha256"`
	SourceRevision *string `json:"sourceRevision"`
	Label          *string `json:"label"`
}

// AddProjectInput adds a PROFILE or SUB binding.
type AddProjectInput struct {
	Role           Role    `json:"role"`
	DTProjectUUID  *string `json:"dtProjectUuid"`
	BOMSerial      *string `json:"bomSerialNumber"`
	BOMVersion     *int    `json:"bomVersion"`
	BOMSHA256      *string `json:"bomSha256"`
	SourceRevision *string `json:"sourceRevision"`
	Label          *string `json:"label"`
}

type ListOptions struct {
	Offset int
	Limit  int
}

func DefaultListOptions() ListOptions {
	return ListOptions{Offset: 0, Limit: 100}
}
