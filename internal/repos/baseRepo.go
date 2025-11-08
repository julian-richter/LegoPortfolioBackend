package repos

import (
	"context"
	"fmt"

	"github.com/charmbracelet/log"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/sync/errgroup"

	"LegoManagerAPI/internal/models"
)

// BaseRepository provides common repository utilities and database access
// specific repositories should embed this and implement their own crud operations
type BaseRepository[T models.Model] struct {
	db        *pgxpool.Pool
	tableName string
}

// NewBaseRepository creates a new BaseRepository
func NewBaseRepository[T models.Model](db *pgxpool.Pool, tableName string) *BaseRepository[T] {
	return &BaseRepository[T]{
		db:        db,
		tableName: tableName,
	}
}

// DB returns the underlying database connection
func (r *BaseRepository[T]) DB() *pgxpool.Pool {
	return r.db
}

// Tablename returns the table name for the model
func (r *BaseRepository[T]) Tablename() string {
	return r.tableName
}

// Count returns the total number of entities in the table
func (r *BaseRepository[T]) Count(ctx context.Context) (int, error) {
	query := fmt.Sprintf("SELECT COUNT(*) FROM %s", r.tableName)
	var count int64
	err := r.db.QueryRow(ctx, query).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("error counting rows: %w", err)
	}

	return int(count), nil
}

// Delete removes an entity by ID
func (r *BaseRepository[T]) Delete(ctx context.Context, id int64) error {
	query := fmt.Sprintf("DELETE FROM %s WHERE id = $1", r.tableName)

	result, err := r.db.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete entity from %s: %w", r.tableName, err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("entity with id %d not found in %s", id, r.tableName)
	}

	log.Debug("Entity deleted", "table", r.tableName, "id", id)
	return nil
}

// Exists checks if an entity with the given ID exists
func (r *BaseRepository[T]) Exists(ctx context.Context, id int64) (bool, error) {
	query := fmt.Sprintf("SELECT EXISTS(SELECT 1 FROM %s WHERE id = $1)", r.tableName)

	var exists bool
	err := r.db.QueryRow(ctx, query, id).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("failed to check existence in %s: %w", r.tableName, err)
	}

	return exists, nil
}

// WithTransaction executes a function within a database transaction
// If the function returns an error, the transactio is rolled back
// Otherwise it's commited
func (r *BaseRepository[T]) WithTransaction(ctx context.Context, fn func(tx pgx.Tx) error) error {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}

	defer func() {
		if p := recover(); p != nil {
			if rbErr := tx.Rollback(ctx); rbErr != nil {
				log.Error("Failed to rollback transaction after panic", "error", rbErr, "panic", p)
			}
			panic(p) // Re-throw panic after rollback
		}
	}()
	if err := fn(tx); err != nil {
		if rbErr := tx.Rollback(ctx); rbErr != nil {
			log.Error("Failed to rollback transaction", "error", rbErr)
			return fmt.Errorf("transaction error: %w (rollback also failed: %v)", err, rbErr)
		}
		return err
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// BatchOperation executes a function for each item concurrenlty using go-routines
// maxConcurrency limits the number of concurrent operations
// This is useful for bulk operations that don't need to be in a transaction
func (r *BaseRepository[T]) BatchOperation(
	ctx context.Context,
	items []T,
	maxConcurrency int,
	operation func(ctx context.Context, item T) error,
) error {
	if len(items) == 0 {
		return nil
	}

	g, gCtx := errgroup.WithContext(ctx)
	sem := make(chan struct{}, maxConcurrency)

	for _, item := range items {
		item := item // Capture loop variable

		g.Go(func() error {
			sem <- struct{}{}        // Acquire semaphore
			defer func() { <-sem }() // Release semaphore

			return operation(gCtx, item)
		})
	}

	if err := g.Wait(); err != nil {
		return fmt.Errorf("batch operation failed: %w", err)
	}

	log.Debug("Batch operation completed", "table", r.tableName, "count", len(items))
	return nil
}

// BatchOperationWithResults executes a function for each item concurrently and collects results
// This is useful when we need to process items and gather their results
func (r *BaseRepository[T]) BatchOperationWithResults(
	ctx context.Context,
	items []T,
	maxConcurrency int,
	operation func(ctx context.Context, item T) (interface{}, error),
) ([]interface{}, error) {
	if len(items) == 0 {
		return []interface{}{}, nil
	}

	g, gCtx := errgroup.WithContext(ctx)
	sem := make(chan struct{}, maxConcurrency)
	results := make([]interface{}, len(items))

	for i, item := range items {
		i, item := i, item // Capture loop variables

		g.Go(func() error {
			sem <- struct{}{}
			defer func() { <-sem }()

			result, err := operation(gCtx, item)
			if err != nil {
				return err
			}

			results[i] = result
			return nil
		})
	}

	if err := g.Wait(); err != nil {
		return nil, fmt.Errorf("batch operation with results failed: %w", err)
	}

	return results, nil
}

// ConcurrentFetch fetches multiple items by IDs concurrently
// The fetch function should retrieve a single item by ID
func (r *BaseRepository[T]) ConcurrentFetch(
	ctx context.Context,
	ids []int64,
	maxConcurrency int,
	fetchFn func(ctx context.Context, id int64) (*T, error),
) ([]*T, error) {
	if len(ids) == 0 {
		return []*T{}, nil
	}

	g, gCtx := errgroup.WithContext(ctx)
	sem := make(chan struct{}, maxConcurrency)
	results := make([]*T, len(ids))

	for i, id := range ids {
		i, id := i, id // Capture loop variables

		g.Go(func() error {
			sem <- struct{}{}
			defer func() { <-sem }()

			item, err := fetchFn(gCtx, id)
			if err != nil {
				return fmt.Errorf("failed to fetch item %d: %w", id, err)
			}

			results[i] = item
			return nil
		})
	}

	if err := g.Wait(); err != nil {
		return nil, err
	}

	log.Debug("Concurrent fetch completed", "table", r.tableName, "count", len(ids))
	return results, nil
}

// BulkDelete deletes multiple entities by IDs concurrently
func (r *BaseRepository[T]) BulkDelete(ctx context.Context, ids []int64, maxConcurrency int) error {
	if len(ids) == 0 {
		return nil
	}

	g, gCtx := errgroup.WithContext(ctx)
	sem := make(chan struct{}, maxConcurrency)

	for _, id := range ids {
		id := id // Capture loop variable

		g.Go(func() error {
			sem <- struct{}{}
			defer func() { <-sem }()

			return r.Delete(gCtx, id)
		})
	}

	if err := g.Wait(); err != nil {
		return fmt.Errorf("bulk delete failed: %w", err)
	}

	log.Info("Bulk delete completed", "table", r.tableName, "count", len(ids))
	return nil
}

// ExecuteInBatches splits a large slice into batches and processes each batch
// This is useful for very large operations to avoid overwhelming the database
func (r *BaseRepository[T]) ExecuteInBatches(
	ctx context.Context,
	items []T,
	batchSize int,
	processBatch func(ctx context.Context, batch []T) error,
) error {
	if len(items) == 0 {
		return nil
	}

	for i := 0; i < len(items); i += batchSize {
		end := i + batchSize
		if end > len(items) {
			end = len(items)
		}

		batch := items[i:end]

		if err := processBatch(ctx, batch); err != nil {
			return fmt.Errorf("failed to process batch %d-%d: %w", i, end, err)
		}

		log.Debug("Batch processed", "table", r.tableName, "range", fmt.Sprintf("%d-%d", i, end))
	}

	log.Info("All batches processed", "table", r.tableName, "total", len(items))
	return nil
}

// Ping checks if the database connection is alive
func (r *BaseRepository[T]) Ping(ctx context.Context) error {
	return r.db.Ping(ctx)
}
