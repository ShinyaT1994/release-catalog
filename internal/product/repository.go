package product

import (
	"context"
	"database/sql"
	"time"
)

// SQLiteRepository implements Repository for SQLite
type SQLiteRepository struct {
	db *sql.DB
}

// NewSQLiteRepository creates a new SQLite product repository
func NewSQLiteRepository(db *sql.DB) *SQLiteRepository {
	return &SQLiteRepository{db: db}
}

func (r *SQLiteRepository) Create(ctx context.Context, p *Product) error {
	_, err := r.db.ExecContext(ctx,
		`INSERT INTO product (id, name, display_name, description, created_at, updated_at)
		 VALUES (?, ?, ?, ?, ?, ?)`,
		p.ID, p.Name, p.DisplayName, p.Description,
		p.CreatedAt.Format(time.RFC3339), p.UpdatedAt.Format(time.RFC3339))
	return err
}

func (r *SQLiteRepository) FindByID(ctx context.Context, id string) (*Product, error) {
	row := r.db.QueryRowContext(ctx,
		`SELECT id, name, display_name, description, created_at, updated_at FROM product WHERE id = ?`, id)
	var p Product
	var createdAt, updatedAt string
	err := row.Scan(&p.ID, &p.Name, &p.DisplayName, &p.Description, &createdAt, &updatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	p.CreatedAt, _ = time.Parse(time.RFC3339, createdAt)
	p.UpdatedAt, _ = time.Parse(time.RFC3339, updatedAt)
	return &p, nil
}

func (r *SQLiteRepository) FindAll(ctx context.Context, opts ListOptions) ([]*Product, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT id, name, display_name, description, created_at, updated_at FROM product ORDER BY created_at DESC LIMIT ? OFFSET ?`,
		opts.Limit, opts.Offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var products []*Product
	for rows.Next() {
		var p Product
		var createdAt, updatedAt string
		if err := rows.Scan(&p.ID, &p.Name, &p.DisplayName, &p.Description, &createdAt, &updatedAt); err != nil {
			return nil, err
		}
		p.CreatedAt, _ = time.Parse(time.RFC3339, createdAt)
		p.UpdatedAt, _ = time.Parse(time.RFC3339, updatedAt)
		products = append(products, &p)
	}
	return products, rows.Err()
}

func (r *SQLiteRepository) Update(ctx context.Context, p *Product) error {
	_, err := r.db.ExecContext(ctx,
		`UPDATE product SET name = ?, display_name = ?, description = ?, updated_at = ? WHERE id = ?`,
		p.Name, p.DisplayName, p.Description, p.UpdatedAt.Format(time.RFC3339), p.ID)
	return err
}

func (r *SQLiteRepository) Delete(ctx context.Context, id string) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM product WHERE id = ?`, id)
	return err
}
