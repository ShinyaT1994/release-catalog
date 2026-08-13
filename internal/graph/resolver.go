package graph

import (
	"context"
	"strings"

	"github.com/ShinyaT1994/release-catalog/internal/dtclient"
	"github.com/google/uuid"
)

// Resolver resolves BOM-Links recursively into a release graph
type Resolver struct {
	dtClient dtclient.Client
	opts     Options
	visited  map[string]bool
	nodes    []Node
	edges    []Edge
	nodeMap  map[string]string // projectUUID -> nodeID
	cycles   int
}

// NewResolver creates a new BOM-Link resolver
func NewResolver(dtClient dtclient.Client, opts Options) *Resolver {
	return &Resolver{
		dtClient: dtClient,
		opts:     opts,
		visited:  make(map[string]bool),
		nodeMap:  make(map[string]string),
	}
}

// Resolve starts BOM-Link resolution from a root project UUID
func (r *Resolver) Resolve(ctx context.Context, rootProjectUUID string) (*ReleaseGraph, error) {
	rootNodeID, _ := r.resolveProject(ctx, rootProjectUUID, 0)

	unresolvedCount := 0
	for _, n := range r.nodes {
		if n.ResolutionStatus != Resolved {
			unresolvedCount++
		}
	}

	return &ReleaseGraph{
		RootNodeID: rootNodeID,
		Nodes:      r.nodes,
		Edges:      r.edges,
		Metadata: Metadata{
			TotalNodes:      len(r.nodes),
			TotalEdges:      len(r.edges),
			MaxDepthReached: false,
			MaxNodesReached: len(r.nodes) >= r.opts.MaxNodes,
			UnresolvedLinks: unresolvedCount,
			CyclesDetected:  r.cycles,
		},
	}, nil
}

func (r *Resolver) resolveProject(ctx context.Context, projectUUID string, depth int) (string, error) {
	if depth > r.opts.MaxDepth || len(r.nodes) >= r.opts.MaxNodes {
		return "", nil
	}

	if r.visited[projectUUID] {
		r.cycles++
		if nodeID, ok := r.nodeMap[projectUUID]; ok {
			return nodeID, nil
		}
		return "", nil
	}
	r.visited[projectUUID] = true

	project, err := r.dtClient.GetProject(ctx, projectUUID)
	if err != nil {
		nodeID := uuid.New().String()
		r.nodes = append(r.nodes, Node{
			ID: nodeID, ProjectUUID: projectUUID,
			ProjectName: "unknown", ProjectVersion: "unknown",
			ResolutionStatus: MissingProject,
		})
		r.nodeMap[projectUUID] = nodeID
		return nodeID, nil
	}

	bom, err := r.dtClient.GetBOM(ctx, projectUUID)
	if err != nil {
		nodeID := uuid.New().String()
		r.nodes = append(r.nodes, Node{
			ID: nodeID, ProjectUUID: projectUUID,
			ProjectName: project.Name, ProjectVersion: project.Version,
			ResolutionStatus: MissingBOM,
		})
		r.nodeMap[projectUUID] = nodeID
		return nodeID, nil
	}

	nodeID := uuid.New().String()
	r.nodes = append(r.nodes, Node{
		ID: nodeID, ProjectUUID: projectUUID,
		ProjectName: project.Name, ProjectVersion: project.Version,
		BOMSerialNumber: bom.SerialNumber, BOMVersion: bom.Version,
		ResolutionStatus: Resolved,
	})
	r.nodeMap[projectUUID] = nodeID

	// Extract and resolve BOM-Links
	links := extractBOMLinks(bom)
	for _, link := range links {
		if link.serialNumber == "" {
			continue
		}
		childNodeID, _ := r.resolveProject(ctx, link.serialNumber, depth+1)
		if childNodeID == "" {
			continue
		}
		edgeStatus := Resolved
		if cn := r.findNode(childNodeID); cn != nil && cn.ResolutionStatus != Resolved {
			edgeStatus = cn.ResolutionStatus
		}
		r.edges = append(r.edges, Edge{
			SourceNodeID: nodeID, TargetNodeID: childNodeID,
			BOMRef: link.bomRef, ResolutionStatus: edgeStatus,
		})
	}

	return nodeID, nil
}

func (r *Resolver) findNode(nodeID string) *Node {
	for i := range r.nodes {
		if r.nodes[i].ID == nodeID {
			return &r.nodes[i]
		}
	}
	return nil
}

// --- BOM-Link parsing ---

type bomLink struct {
	serialNumber string
	version      int
	bomRef       string
}

func extractBOMLinks(bom *dtclient.CycloneDXBOM) []bomLink {
	var links []bomLink
	for _, comp := range bom.Components {
		for _, ref := range comp.ExternalReferences {
			if ref.Type == "bom" && strings.HasPrefix(ref.URL, "urn:cdx:") {
				if l, ok := parseBOMLink(ref.URL); ok {
					links = append(links, l)
				}
			}
		}
	}
	return links
}

func parseBOMLink(uri string) (bomLink, bool) {
	rest := strings.TrimPrefix(uri, "urn:cdx:")
	if rest == uri {
		return bomLink{}, false
	}

	var link bomLink
	link.version = 1

	parts := strings.SplitN(rest, "#", 2)
	if len(parts) == 2 {
		link.bomRef = parts[1]
	}

	serialAndVersion := strings.SplitN(parts[0], "/", 2)
	link.serialNumber = serialAndVersion[0]
	if len(serialAndVersion) == 2 {
		v := 0
		for _, ch := range serialAndVersion[1] {
			if ch < '0' || ch > '9' {
				return bomLink{}, false
			}
			v = v*10 + int(ch-'0')
		}
		link.version = v
	}

	return link, true
}
