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

// BranchLine represents a branch (Main or Release) entity.
// A release branch forks from a specific main VERSION (forkedFromVersionId).
type BranchLine struct {
	ID                  string       `json:"id"`
	ProductID           string       `json:"productId"`
	Type                BranchType   `json:"type"`
	Name                string       `json:"name"`
	DisplayName         string       `json:"displayName"`
	SourceBranchLineID  *string      `json:"sourceBranchLineId,omitempty"`
	ForkedFromVersionID *string      `json:"forkedFromVersionId,omitempty"`
	Status              BranchStatus `json:"status"`
	CreatedAt           time.Time    `json:"createdAt"`
	UpdatedAt           time.Time    `json:"updatedAt"`
	ClosedAt            *time.Time   `json:"closedAt,omitempty"`
}
