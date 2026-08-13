package graph

import (
	"context"

	"github.com/ShinyaT1994/release-catalog/internal/dtclient"
	"github.com/ShinyaT1994/release-catalog/internal/shared/apperror"
)

// Service implements UseCase for Graph
type Service struct {
	dtClient dtclient.Client
	csFinder BranchCurrentStateFinder
	snapFinder SnapshotFinder
}

func NewService(dtClient dtclient.Client, csFinder BranchCurrentStateFinder, snapFinder SnapshotFinder) *Service {
	return &Service{dtClient: dtClient, csFinder: csFinder, snapFinder: snapFinder}
}

func (s *Service) GetBranchCurrentGraph(ctx context.Context, branchID string, opts Options) (*ReleaseGraph, error) {
	exists, err := s.csFinder.BranchExists(ctx, branchID)
	if err != nil {
		return nil, apperror.New(apperror.CodeInternalError, err.Error())
	}
	if !exists {
		return nil, apperror.New(apperror.CodeBranchNotFound, "branch not found")
	}

	cs, err := s.csFinder.FindByBranchID(ctx, branchID)
	if err != nil {
		return nil, apperror.New(apperror.CodeInternalError, err.Error())
	}
	if cs == nil || cs.RootDTProjectUUID == nil {
		return nil, apperror.New(apperror.CodeInvalidRequest, "no root project configured for this branch")
	}

	resolver := NewResolver(s.dtClient, opts)
	return resolver.Resolve(ctx, *cs.RootDTProjectUUID)
}

func (s *Service) GetReleaseGraph(ctx context.Context, releaseID string, opts Options) (*ReleaseGraph, error) {
	snap, err := s.snapFinder.FindByID(ctx, releaseID)
	if err != nil {
		return nil, apperror.New(apperror.CodeInternalError, err.Error())
	}
	if snap == nil {
		return nil, apperror.New(apperror.CodeReleaseNotFound, "release not found")
	}
	if snap.RootDTProjectUUID == nil {
		return nil, apperror.New(apperror.CodeInvalidRequest, "no root project configured for this release")
	}

	resolver := NewResolver(s.dtClient, opts)
	return resolver.Resolve(ctx, *snap.RootDTProjectUUID)
}
