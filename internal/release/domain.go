package release

import "time"

// SnapshotType represents the type of snapshot
type SnapshotType string

const (
	TypeMainSnapshot SnapshotType = "MAIN_SNAPSHOT"
	TypeRelease      SnapshotType = "RELEASE"
)

// Status represents the release lifecycle status
type Status string

const (
	StatusDraft      Status = "draft"
	StatusTesting    Status = "testing"
	StatusApproved   Status = "approved"
	StatusReleased   Status = "released"
	StatusDeprecated Status = "deprecated"
	StatusEOS        Status = "end_of_support"
)

// Snapshot represents a release snapshot entity
type Snapshot struct {
	ID                string       `json:"id"`
	BranchLineID      string       `json:"branchLineId"`
	SnapshotType      SnapshotType `json:"snapshotType"`
	Version           string       `json:"version"`
	Status            Status       `json:"status"`
	RootDTProjectUUID *string      `json:"rootDtProjectUuid,omitempty"`
	RootBOMSerial     *string      `json:"rootBomSerialNumber,omitempty"`
	RootBOMVersion    *int         `json:"rootBomVersion,omitempty"`
	RootBOMSHA256     *string      `json:"rootBomSha256,omitempty"`
	SourceRevision    *string      `json:"sourceRevision,omitempty"`
	CreatedAt         time.Time    `json:"createdAt"`
	ReleasedAt        *time.Time   `json:"releasedAt,omitempty"`
}
