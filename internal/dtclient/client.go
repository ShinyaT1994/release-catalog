package dtclient

import "context"

// Client defines the interface for Dependency-Track operations (Driven Port)
type Client interface {
	GetProject(ctx context.Context, uuid string) (*Project, error)
	ProjectExists(ctx context.Context, uuid string) (bool, error)
	GetBOM(ctx context.Context, projectUUID string) (*CycloneDXBOM, error)
	GetVulnerabilities(ctx context.Context, projectUUID string) ([]*Vulnerability, error)
}

// --- DT Models ---

type Project struct {
	UUID    string `json:"uuid"`
	Name    string `json:"name"`
	Version string `json:"version"`
}

type CycloneDXBOM struct {
	BOMFormat    string         `json:"bomFormat"`
	SpecVersion  string         `json:"specVersion"`
	SerialNumber string         `json:"serialNumber"`
	Version      int            `json:"version"`
	Components   []BOMComponent `json:"components"`
}

type BOMComponent struct {
	Type               string        `json:"type"`
	Name               string        `json:"name"`
	Version            string        `json:"version"`
	BOMRef             string        `json:"bom-ref"`
	ExternalReferences []ExternalRef `json:"externalReferences,omitempty"`
}

type ExternalRef struct {
	Type string `json:"type"`
	URL  string `json:"url"`
}

type Vulnerability struct {
	UUID     string   `json:"uuid"`
	VulnID   string   `json:"vulnId"`
	Source   string   `json:"source"`
	Severity string   `json:"severity"`
}
