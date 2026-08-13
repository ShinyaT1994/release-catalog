package dtclient

import (
	"context"
	"fmt"

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

func (s *StubClient) loadSampleData() {
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
