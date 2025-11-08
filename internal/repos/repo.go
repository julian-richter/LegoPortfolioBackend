package repos

import (
	"context"

	"LegoManagerAPI/internal/models"
)

// Repository is the base interface for all repos
type Repository[T models.Model] interface {
	// Basic CRUD
	Create(ctx context.Context, entity *T) error
	FindByID(ctx context.Context, id int64) (*T, error)
	Update(ctx context.Context, entity *T) error
	Delete(ctx context.Context, id int64) error
	List(ctx context.Context, limit, offset int) ([]*T, error)
	Count(ctx context.Context) (int, error)

	// Batch operations (uses goroutines internally)
	CreateBatch(ctx context.Context, entities []*T) error
	FindByIDs(ctx context.Context, ids []int64) ([]*T, error)
}

// Tablenamer is an interface for models that can return their table name
type Tablenamer interface {
	Tablename() string
}
