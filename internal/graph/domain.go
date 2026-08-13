package graph

// ResolutionStatus represents the resolution state of a BOM-Link
type ResolutionStatus string

const (
	Resolved       ResolutionStatus = "resolved"
	MissingProject ResolutionStatus = "missing_project"
	MissingBOM     ResolutionStatus = "missing_bom"
	MissingBOMRef  ResolutionStatus = "missing_bom_ref"
	Invalid        ResolutionStatus = "invalid"
)

// Node represents a node in the release graph
type Node struct {
	ID               string           `json:"id"`
	ProjectUUID      string           `json:"projectUUID"`
	ProjectName      string           `json:"projectName"`
	ProjectVersion   string           `json:"projectVersion"`
	BOMSerialNumber  string           `json:"bomSerialNumber,omitempty"`
	BOMVersion       int              `json:"bomVersion,omitempty"`
	ResolutionStatus ResolutionStatus `json:"resolutionStatus"`
}

// Edge represents an edge in the release graph
type Edge struct {
	SourceNodeID     string           `json:"sourceNodeId"`
	TargetNodeID     string           `json:"targetNodeId"`
	BOMRef           string           `json:"bomRef,omitempty"`
	ResolutionStatus ResolutionStatus `json:"resolutionStatus"`
}

// ReleaseGraph represents the complete release graph
type ReleaseGraph struct {
	RootNodeID string   `json:"rootNodeId"`
	Nodes      []Node   `json:"nodes"`
	Edges      []Edge   `json:"edges"`
	Metadata   Metadata `json:"metadata"`
}

// Metadata provides summary info about the graph
type Metadata struct {
	TotalNodes      int  `json:"totalNodes"`
	TotalEdges      int  `json:"totalEdges"`
	MaxDepthReached bool `json:"maxDepthReached"`
	MaxNodesReached bool `json:"maxNodesReached"`
	UnresolvedLinks int  `json:"unresolvedLinks"`
	CyclesDetected  int  `json:"cyclesDetected"`
}

// Options controls graph traversal
type Options struct {
	MaxDepth int
	MaxNodes int
}

func DefaultOptions() Options {
	return Options{MaxDepth: 10, MaxNodes: 1000}
}
