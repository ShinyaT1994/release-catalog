package branch

import "time"

// BranchType represents the type of branch line
type BranchType string

const (
	TypeMain    BranchType = "MAIN"
	TypeRelease BranchType = "RELEASE"
)

// BranchStatus represents the lifecycle status of a branch
type BranchStatus string

const (
	StatusActive       BranchStatus = "active"
	StatusMaintenance  BranchStatus = "maintenance"
	StatusSecurityOnly BranchStatus = "security_only"
	StatusEOS          BranchStatus = "end_of_support"
	StatusClosed       BranchStatus = "closed"
)

// BranchLine represents a branch (Main or Release) entity
type BranchLine struct {
	ID                   string       `json:"id"`
	ProductID            string       `json:"productId"`
	Type                 BranchType   `json:"type"`
	Name                 string       `json:"name"`
	DisplayName          string       `json:"displayName"`
	SourceBranchLineID   *string      `json:"sourceBranchLineId,omitempty"`
	ForkedFromSnapshotID *string      `json:"forkedFromSnapshotId,omitempty"`
	Status               BranchStatus `json:"status"`
	CreatedAt            time.Time    `json:"createdAt"`
	UpdatedAt            time.Time    `json:"updatedAt"`
	ClosedAt             *time.Time   `json:"closedAt,omitempty"`
}

// CurrentState represents the current state of a branch
type CurrentState struct {
	BranchLineID      string    `json:"branchLineId"`
	RootDTProjectUUID *string   `json:"rootDtProjectUuid,omitempty"`
	RootBOMSerial     *string   `json:"rootBomSerialNumber,omitempty"`
	RootBOMVersion    *int      `json:"rootBomVersion,omitempty"`
	RootBOMSHA256     *string   `json:"rootBomSha256,omitempty"`
	SourceRevision    *string   `json:"sourceRevision,omitempty"`
	UpdatedAt         time.Time `json:"updatedAt"`
}
