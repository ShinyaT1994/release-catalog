package graph

import (
	"context"

	"github.com/ShinyaT1994/release-catalog/internal/dtclient"
	"github.com/ShinyaT1994/release-catalog/internal/shared/apperror"
)

// SBOMService resolves the BOM-Link SBOM graph for a role-tagged project.
type SBOMService struct {
	dtClient dtclient.Client
	finder   VersionProjectFinder
}

func NewSBOMService(dtClient dtclient.Client, finder VersionProjectFinder) *SBOMService {
	return &SBOMService{dtClient: dtClient, finder: finder}
}

func (s *SBOMService) GetProjectGraph(ctx context.Context, versionID, role string, opts Options) (*ReleaseGraph, error) {
	exists, err := s.finder.VersionExists(ctx, versionID)
	if err != nil {
		return nil, apperror.New(apperror.CodeInternalError, err.Error())
	}
	if !exists {
		return nil, apperror.New(apperror.CodeVersionNotFound, "version not found")
	}

	uuid, err := s.finder.FindProjectUUID(ctx, versionID, role)
	if err != nil {
		return nil, apperror.New(apperror.CodeInternalError, err.Error())
	}
	if uuid == nil || *uuid == "" {
		return nil, apperror.New(apperror.CodeInvalidRequest, "no DT project configured for this role")
	}

	resolver := NewResolver(s.dtClient, opts)
	return resolver.Resolve(ctx, *uuid)
}
