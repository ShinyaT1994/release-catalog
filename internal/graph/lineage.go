package graph

import (
	"context"
	"sort"

	"github.com/ShinyaT1994/release-catalog/internal/shared/apperror"
)

// LineageService builds the version-lineage graph and timelines from RC DB only
// (no Dependency-Track calls).
type LineageService struct {
	src LineageSource
}

func NewLineageService(src LineageSource) *LineageService {
	return &LineageService{src: src}
}

func (s *LineageService) GetProductLineage(ctx context.Context, productID string, axis LineageAxis) (*LineageGraph, error) {
	exists, err := s.src.ProductExists(ctx, productID)
	if err != nil {
		return nil, apperror.New(apperror.CodeInternalError, err.Error())
	}
	if !exists {
		return nil, apperror.New(apperror.CodeProductNotFound, "product not found")
	}

	if axis != AxisVersion && axis != AxisReleaseDate {
		axis = AxisReleaseDate
	}

	rows, err := s.src.ListVersionsByProduct(ctx, productID)
	if err != nil {
		return nil, apperror.New(apperror.CodeInternalError, err.Error())
	}

	sortRows(rows, axis)

	present := make(map[string]bool, len(rows))
	for _, r := range rows {
		present[r.VersionID] = true
	}

	g := &LineageGraph{ProductID: productID, Axis: axis, Nodes: []LineageNode{}, Edges: []LineageEdge{}}
	for _, r := range rows {
		g.Nodes = append(g.Nodes, LineageNode{
			VersionID:     r.VersionID,
			BranchLineID:  r.BranchLineID,
			BranchName:    r.BranchName,
			BranchType:    r.BranchType,
			VersionString: r.VersionString,
			Status:        r.Status,
			ReleaseDate:   r.ReleaseDate,
			CreatedAt:     r.CreatedAt,
		})
		if r.ParentVersionID != nil && present[*r.ParentVersionID] {
			g.Edges = append(g.Edges, LineageEdge{
				SourceVersionID: *r.ParentVersionID,
				TargetVersionID: r.VersionID,
				Kind:            "parent",
			})
		}
		if r.ForkedFromVersionID != nil && present[*r.ForkedFromVersionID] {
			g.Edges = append(g.Edges, LineageEdge{
				SourceVersionID: *r.ForkedFromVersionID,
				TargetVersionID: r.VersionID,
				Kind:            "fork",
			})
		}
	}
	return g, nil
}

func (s *LineageService) GetProductTimeline(ctx context.Context, productID string) (*Timeline, error) {
	exists, err := s.src.ProductExists(ctx, productID)
	if err != nil {
		return nil, apperror.New(apperror.CodeInternalError, err.Error())
	}
	if !exists {
		return nil, apperror.New(apperror.CodeProductNotFound, "product not found")
	}
	rows, err := s.src.ListVersionsByProduct(ctx, productID)
	if err != nil {
		return nil, apperror.New(apperror.CodeInternalError, err.Error())
	}
	return toTimeline(rows), nil
}

func (s *LineageService) GetBranchTimeline(ctx context.Context, branchID string) (*Timeline, error) {
	exists, err := s.src.BranchExists(ctx, branchID)
	if err != nil {
		return nil, apperror.New(apperror.CodeInternalError, err.Error())
	}
	if !exists {
		return nil, apperror.New(apperror.CodeBranchNotFound, "branch not found")
	}
	rows, err := s.src.ListVersionsByBranch(ctx, branchID)
	if err != nil {
		return nil, apperror.New(apperror.CodeInternalError, err.Error())
	}
	return toTimeline(rows), nil
}

func toTimeline(rows []VersionRow) *Timeline {
	sortRows(rows, AxisReleaseDate)
	t := &Timeline{Entries: []TimelineEntry{}}
	for _, r := range rows {
		t.Entries = append(t.Entries, TimelineEntry{
			VersionID:     r.VersionID,
			BranchLineID:  r.BranchLineID,
			BranchName:    r.BranchName,
			BranchType:    r.BranchType,
			VersionString: r.VersionString,
			Status:        r.Status,
			ReleaseDate:   r.ReleaseDate,
			CreatedAt:     r.CreatedAt,
		})
	}
	return t
}

// sortRows orders rows for stable output. For the releaseDate axis, rows with a
// release date sort first (chronologically), then rows without a date by
// createdAt. For the version axis, rows sort by branch then createdAt.
func sortRows(rows []VersionRow, axis LineageAxis) {
	sort.SliceStable(rows, func(i, j int) bool {
		a, b := rows[i], rows[j]
		if axis == AxisReleaseDate {
			ad, bd := a.ReleaseDate, b.ReleaseDate
			switch {
			case ad != nil && bd != nil:
				if *ad != *bd {
					return *ad < *bd
				}
			case ad != nil && bd == nil:
				return true
			case ad == nil && bd != nil:
				return false
			}
			return a.CreatedAt < b.CreatedAt
		}
		// version axis: group by branch, then chronological within branch
		if a.BranchLineID != b.BranchLineID {
			return a.BranchLineID < b.BranchLineID
		}
		return a.CreatedAt < b.CreatedAt
	})
}
