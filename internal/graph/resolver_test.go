package graph_test

import (
	"context"
	"testing"

	"github.com/ShinyaT1994/release-catalog/internal/dtclient"
	"github.com/ShinyaT1994/release-catalog/internal/graph"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestResolveGraph_BasicTree(t *testing.T) {
	stub := dtclient.NewStubClient()
	opts := graph.DefaultOptions()

	resolver := graph.NewResolver(stub, opts)
	g, err := resolver.Resolve(context.Background(), "00000000-0000-0000-0000-000000000001")

	require.NoError(t, err)
	assert.NotEmpty(t, g.RootNodeID)
	// Root -> Backend, Frontend; Backend -> Common; Frontend -> Common
	assert.Equal(t, 4, g.Metadata.TotalNodes)
	assert.Equal(t, 4, g.Metadata.TotalEdges) // root->backend, root->frontend, backend->common, frontend->common(cycle reuse)
	assert.Equal(t, 0, g.Metadata.UnresolvedLinks)
}

func TestResolveGraph_CycleDetection(t *testing.T) {
	stub := &dtclient.StubClient{
		Projects: map[string]*dtclient.Project{
			"a": {UUID: "a", Name: "A", Version: "1.0"},
			"b": {UUID: "b", Name: "B", Version: "1.0"},
		},
		BOMs: map[string]*dtclient.CycloneDXBOM{
			"a": {
				BOMFormat: "CycloneDX", SpecVersion: "1.5", SerialNumber: "urn:uuid:a", Version: 1,
				Components: []dtclient.BOMComponent{
					{Name: "B", Version: "1.0", BOMRef: "b-ref",
						ExternalReferences: []dtclient.ExternalRef{{Type: "bom", URL: "urn:cdx:b/1"}}},
				},
			},
			"b": {
				BOMFormat: "CycloneDX", SpecVersion: "1.5", SerialNumber: "urn:uuid:b", Version: 1,
				Components: []dtclient.BOMComponent{
					{Name: "A", Version: "1.0", BOMRef: "a-ref",
						ExternalReferences: []dtclient.ExternalRef{{Type: "bom", URL: "urn:cdx:a/1"}}},
				},
			},
		},
	}

	resolver := graph.NewResolver(stub, graph.DefaultOptions())
	g, err := resolver.Resolve(context.Background(), "a")

	require.NoError(t, err)
	assert.Equal(t, 2, g.Metadata.TotalNodes)
	assert.True(t, g.Metadata.CyclesDetected > 0)
}

func TestResolveGraph_MaxDepth(t *testing.T) {
	stub := dtclient.NewStubClient()
	opts := graph.Options{MaxDepth: 1, MaxNodes: 1000}

	resolver := graph.NewResolver(stub, opts)
	g, err := resolver.Resolve(context.Background(), "00000000-0000-0000-0000-000000000001")

	require.NoError(t, err)
	// At depth 1 we only get root + direct children (Backend, Frontend) but not Common
	assert.True(t, g.Metadata.TotalNodes <= 3)
}

func TestResolveGraph_MissingProject(t *testing.T) {
	stub := &dtclient.StubClient{
		Projects: map[string]*dtclient.Project{
			"root": {UUID: "root", Name: "Root", Version: "1.0"},
		},
		BOMs: map[string]*dtclient.CycloneDXBOM{
			"root": {
				BOMFormat: "CycloneDX", SpecVersion: "1.5", SerialNumber: "urn:uuid:root", Version: 1,
				Components: []dtclient.BOMComponent{
					{Name: "Missing", Version: "1.0", BOMRef: "missing-ref",
						ExternalReferences: []dtclient.ExternalRef{{Type: "bom", URL: "urn:cdx:nonexistent/1"}}},
				},
			},
		},
	}

	resolver := graph.NewResolver(stub, graph.DefaultOptions())
	g, err := resolver.Resolve(context.Background(), "root")

	require.NoError(t, err)
	assert.Equal(t, 2, g.Metadata.TotalNodes)
	assert.Equal(t, 1, g.Metadata.UnresolvedLinks)
}

func TestResolveGraph_SharedDependency(t *testing.T) {
	stub := dtclient.NewStubClient()
	opts := graph.DefaultOptions()

	resolver := graph.NewResolver(stub, opts)
	g, err := resolver.Resolve(context.Background(), "00000000-0000-0000-0000-000000000001")

	require.NoError(t, err)
	// Common is shared by Backend and Frontend - should appear once as node
	commonCount := 0
	for _, n := range g.Nodes {
		if n.ProjectName == "Common" {
			commonCount++
		}
	}
	assert.Equal(t, 1, commonCount)
}
