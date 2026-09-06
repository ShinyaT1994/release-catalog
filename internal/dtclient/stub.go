package dtclient

import (
	"context"
	"fmt"
	"strings"

	"github.com/ShinyaT1994/release-catalog/internal/shared/apperror"
)

// StubClient is a development/test stub for the DT client
type StubClient struct {
	Projects map[string]*Project
	BOMs     map[string]*CycloneDXBOM
}

// NewStubClient creates a StubClient with sample data
func NewStubClient() *StubClient {
	s := &StubClient{
		Projects: make(map[string]*Project),
		BOMs:     make(map[string]*CycloneDXBOM),
	}
	s.loadSampleData()
	return s
}

func (s *StubClient) GetProject(ctx context.Context, uuid string) (*Project, error) {
	p, ok := s.Projects[uuid]
	if !ok {
		return nil, apperror.New(apperror.CodeRootProjectNotFound, fmt.Sprintf("project %s not found in Dependency-Track", uuid))
	}
	return p, nil
}

func (s *StubClient) ProjectExists(ctx context.Context, uuid string) (bool, error) {
	_, ok := s.Projects[uuid]
	return ok, nil
}

func (s *StubClient) GetBOM(ctx context.Context, projectUUID string) (*CycloneDXBOM, error) {
	bom, ok := s.BOMs[projectUUID]
	if !ok {
		return nil, apperror.New(apperror.CodeRootProjectNotFound, fmt.Sprintf("BOM not found for project %s", projectUUID))
	}
	return bom, nil
}

func (s *StubClient) GetVulnerabilities(ctx context.Context, projectUUID string) ([]*Vulnerability, error) {
	return []*Vulnerability{}, nil
}

func (s *StubClient) ListProjects(ctx context.Context) ([]*Project, error) {
	projects := make([]*Project, 0, len(s.Projects))
	for _, p := range s.Projects {
		projects = append(projects, p)
	}
	return projects, nil
}

func (s *StubClient) SearchProjects(ctx context.Context, name string) ([]*Project, error) {
	var projects []*Project
	for _, p := range s.Projects {
		if containsIgnoreCase(p.Name, name) {
			projects = append(projects, p)
		}
	}
	return projects, nil
}

func (s *StubClient) loadSampleData() {
	// Primary demo: TopBOM -> SubBOM (with explicit CycloneDX versions)
	s.Projects["00000000-0000-0000-0000-0000000000a1"] = &Project{UUID: "00000000-0000-0000-0000-0000000000a1", Name: "TopBOM", Version: "1.0.0"}
	s.Projects["00000000-0000-0000-0000-0000000000a2"] = &Project{UUID: "00000000-0000-0000-0000-0000000000a2", Name: "SubBOM", Version: "2.1.0"}

	s.BOMs["00000000-0000-0000-0000-0000000000a1"] = &CycloneDXBOM{
		BOMFormat: "CycloneDX", SpecVersion: "1.5",
		SerialNumber: "urn:uuid:00000000-0000-0000-0000-0000000000a1", Version: 3,
		Components: []BOMComponent{
			{Type: "application", Name: "SubBOM", Version: "2.1.0", BOMRef: "sub-bom-ref",
				ExternalReferences: []ExternalRef{{Type: "bom", URL: "urn:cdx:00000000-0000-0000-0000-0000000000a2/5"}}},
		},
	}
	s.Projects["00000000-0000-0000-0000-0000000000a3"] = &Project{UUID: "00000000-0000-0000-0000-0000000000a3", Name: "NestedBOM", Version: "0.9.0"}

	s.BOMs["00000000-0000-0000-0000-0000000000a2"] = &CycloneDXBOM{
		BOMFormat: "CycloneDX", SpecVersion: "1.5",
		SerialNumber: "urn:uuid:00000000-0000-0000-0000-0000000000a2", Version: 6,
		Components: []BOMComponent{
			{Type: "application", Name: "NestedBOM", Version: "0.9.0", BOMRef: "nested-bom-ref",
				ExternalReferences: []ExternalRef{{Type: "bom", URL: "urn:cdx:00000000-0000-0000-0000-0000000000a3/2"}}},
			{Type: "library", Name: "example-lib", Version: "1.2.3", BOMRef: "pkg:example/example-lib@1.2.3"},
		},
	}
	s.BOMs["00000000-0000-0000-0000-0000000000a3"] = &CycloneDXBOM{
		BOMFormat: "CycloneDX", SpecVersion: "1.5",
		SerialNumber: "urn:uuid:00000000-0000-0000-0000-0000000000a3", Version: 2,
		Components: []BOMComponent{
			{Type: "library", Name: "nested-lib", Version: "0.1.0", BOMRef: "pkg:example/nested-lib@0.1.0"},
		},
	}

	// Multi-level sample used by graph unit tests
	s.Projects["00000000-0000-0000-0000-000000000001"] = &Project{UUID: "00000000-0000-0000-0000-000000000001", Name: "ProductRoot", Version: "8.0.0"}
	s.Projects["00000000-0000-0000-0000-000000000002"] = &Project{UUID: "00000000-0000-0000-0000-000000000002", Name: "Backend", Version: "8.0.0"}
	s.Projects["00000000-0000-0000-0000-000000000003"] = &Project{UUID: "00000000-0000-0000-0000-000000000003", Name: "Frontend", Version: "6.0.0"}
	s.Projects["00000000-0000-0000-0000-000000000004"] = &Project{UUID: "00000000-0000-0000-0000-000000000004", Name: "Common", Version: "4.0.0"}

	s.BOMs["00000000-0000-0000-0000-000000000001"] = &CycloneDXBOM{
		BOMFormat: "CycloneDX", SpecVersion: "1.5",
		SerialNumber: "urn:uuid:00000000-0000-0000-0000-000000000001", Version: 1,
		Components: []BOMComponent{
			{Type: "application", Name: "Backend", Version: "8.0.0", BOMRef: "backend-ref",
				ExternalReferences: []ExternalRef{{Type: "bom", URL: "urn:cdx:00000000-0000-0000-0000-000000000002/1"}}},
			{Type: "application", Name: "Frontend", Version: "6.0.0", BOMRef: "frontend-ref",
				ExternalReferences: []ExternalRef{{Type: "bom", URL: "urn:cdx:00000000-0000-0000-0000-000000000003/1"}}},
		},
	}
	s.BOMs["00000000-0000-0000-0000-000000000002"] = &CycloneDXBOM{
		BOMFormat: "CycloneDX", SpecVersion: "1.5",
		SerialNumber: "urn:uuid:00000000-0000-0000-0000-000000000002", Version: 1,
		Components: []BOMComponent{
			{Type: "library", Name: "Common", Version: "4.0.0", BOMRef: "common-ref",
				ExternalReferences: []ExternalRef{{Type: "bom", URL: "urn:cdx:00000000-0000-0000-0000-000000000004/1"}}},
		},
	}
	s.BOMs["00000000-0000-0000-0000-000000000003"] = &CycloneDXBOM{
		BOMFormat: "CycloneDX", SpecVersion: "1.5",
		SerialNumber: "urn:uuid:00000000-0000-0000-0000-000000000003", Version: 1,
		Components: []BOMComponent{
			{Type: "library", Name: "Common", Version: "4.0.0", BOMRef: "common-ref",
				ExternalReferences: []ExternalRef{{Type: "bom", URL: "urn:cdx:00000000-0000-0000-0000-000000000004/1"}}},
		},
	}
	s.BOMs["00000000-0000-0000-0000-000000000004"] = &CycloneDXBOM{
		BOMFormat: "CycloneDX", SpecVersion: "1.5",
		SerialNumber: "urn:uuid:00000000-0000-0000-0000-000000000004", Version: 1,
		Components: []BOMComponent{},
	}
}

func containsIgnoreCase(s, substr string) bool {
	return strings.Contains(strings.ToLower(s), strings.ToLower(substr))
}
