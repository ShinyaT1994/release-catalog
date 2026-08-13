package product

import "context"

// --- Driving Port (UseCase interface) ---

// UseCase defines the business operations for Product
type UseCase interface {
	Create(ctx context.Context, input CreateInput) (*Product, error)
	GetByID(ctx context.Context, id string) (*Product, error)
	List(ctx context.Context, opts ListOptions) ([]*Product, error)
	Update(ctx context.Context, id string, input UpdateInput) (*Product, error)
	Delete(ctx context.Context, id string) error
}

// --- Driven Port (Repository interface) ---

// Repository defines the persistence operations for Product
type Repository interface {
	Create(ctx context.Context, p *Product) error
	FindByID(ctx context.Context, id string) (*Product, error)
	FindAll(ctx context.Context, opts ListOptions) ([]*Product, error)
	Update(ctx context.Context, p *Product) error
	Delete(ctx context.Context, id string) error
}

// --- DTOs ---

type CreateInput struct {
	Name        string `json:"name"`
	DisplayName string `json:"displayName"`
	Description string `json:"description"`
}

type UpdateInput struct {
	DisplayName *string `json:"displayName"`
	Description *string `json:"description"`
}

type ListOptions struct {
	Offset int
	Limit  int
}

func DefaultListOptions() ListOptions {
	return ListOptions{Offset: 0, Limit: 100}
}
