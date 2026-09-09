package version

import "time"

// Status represents the version lifecycle status.
type Status string

const (
	StatusIncomplete Status = "incomplete"
	StatusFinalized  Status = "finalized"
)

// Role represents the role of a DT project bound to a version.
type Role string

const (
	RoleRoot    Role = "ROOT"
	RoleProfile Role = "PROFILE"
	RoleSub     Role = "SUB"
)

// Version is the git-like commit on a branch line. It replaces the old
// snapshot and branch_current_state model.
type Version struct {
	ID                   string     `json:"id"`
	BranchLineID         string     `json:"branchLineId"`
	VersionString        string     `json:"versionString"`
	Status               Status     `json:"status"`
	ParentVersionID      *string    `json:"parentVersionId,omitempty"`
	ForkedFromVersionID  *string    `json:"forkedFromVersionId,omitempty"`
	Location             *string    `json:"location,omitempty"`
	Customer             *string    `json:"customer,omitempty"`
	ReleaseDate          *time.Time `json:"releaseDate,omitempty"`
	CreatedAt            time.Time  `json:"createdAt"`
	UpdatedAt            time.Time  `json:"updatedAt"`
	// Projects is populated on reads that hydrate role-tagged bindings.
	Projects []*DTProject `json:"projects,omitempty"`
}

// DTProject is a role-tagged Dependency-Track project binding for a version.
type DTProject struct {
	ID             int64     `json:"id"`
	VersionID      string    `json:"versionId"`
	Role           Role      `json:"role"`
	DTProjectUUID  *string   `json:"dtProjectUuid,omitempty"`
	BOMSerial      *string   `json:"bomSerialNumber,omitempty"`
	BOMVersion     *int      `json:"bomVersion,omitempty"`
	BOMSHA256      *string   `json:"bomSha256,omitempty"`
	SourceRevision *string   `json:"sourceRevision,omitempty"`
	Label          *string   `json:"label,omitempty"`
	CreatedAt      time.Time `json:"createdAt"`
	UpdatedAt      time.Time `json:"updatedAt"`
}
