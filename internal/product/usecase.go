package product

import (
	"context"
	"time"

	"github.com/ShinyaT1994/release-catalog/internal/branch"
	"github.com/ShinyaT1994/release-catalog/internal/shared/apperror"
	"github.com/google/uuid"
)

// Service implements UseCase
type Service struct {
	repo       Repository
	branchRepo branch.Repository
}

// NewService creates a new product service
func NewService(repo Repository, branchRepo branch.Repository) *Service {
	return &Service{repo: repo, branchRepo: branchRepo}
}

func (s *Service) Create(ctx context.Context, input CreateInput) (*Product, error) {
	if input.Name == "" {
		return nil, apperror.New(apperror.CodeInvalidRequest, "name is required")
	}

	now := time.Now().UTC()
	p := &Product{
		ID:          uuid.New().String(),
		Name:        input.Name,
		DisplayName: input.DisplayName,
		Description: input.Description,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	if err := s.repo.Create(ctx, p); err != nil {
		return nil, apperror.New(apperror.CodeInternalError, "failed to create product: "+err.Error())
	}

	// Auto-create Main Line
	mainBranch := &branch.BranchLine{
		ID:          uuid.New().String(),
		ProductID:   p.ID,
		Type:        branch.TypeMain,
		Name:        "main",
		DisplayName: "Main",
		Status:      branch.StatusActive,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	if err := s.branchRepo.Create(ctx, mainBranch); err != nil {
		return nil, apperror.New(apperror.CodeInternalError, "failed to create main branch: "+err.Error())
	}

	return p, nil
}

func (s *Service) GetByID(ctx context.Context, id string) (*Product, error) {
	p, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, apperror.New(apperror.CodeInternalError, err.Error())
	}
	if p == nil {
		return nil, apperror.New(apperror.CodeProductNotFound, "product not found")
	}
	return p, nil
}

func (s *Service) List(ctx context.Context, opts ListOptions) ([]*Product, error) {
	return s.repo.FindAll(ctx, opts)
}

func (s *Service) Update(ctx context.Context, id string, input UpdateInput) (*Product, error) {
	p, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, apperror.New(apperror.CodeInternalError, err.Error())
	}
	if p == nil {
		return nil, apperror.New(apperror.CodeProductNotFound, "product not found")
	}

	if input.DisplayName != nil {
		p.DisplayName = *input.DisplayName
	}
	if input.Description != nil {
		p.Description = *input.Description
	}
	p.UpdatedAt = time.Now().UTC()

	if err := s.repo.Update(ctx, p); err != nil {
		return nil, apperror.New(apperror.CodeInternalError, err.Error())
	}
	return p, nil
}

func (s *Service) Delete(ctx context.Context, id string) error {
	p, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return apperror.New(apperror.CodeInternalError, err.Error())
	}
	if p == nil {
		return apperror.New(apperror.CodeProductNotFound, "product not found")
	}
	return s.repo.Delete(ctx, id)
}
