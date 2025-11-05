package repos

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/sync/errgroup"

	"LegoManagerAPI/internal/models"
)

// UserRepository handles user data operations
type UserRepository struct {
	*BaseRepository[models.User] // Non-pointer generic
}

// NewUserRepository creates a new User repository
func NewUserRepository(db *pgxpool.Pool) *UserRepository {
	return &UserRepository{
		BaseRepository: NewBaseRepository[models.User](db, "users"),
	}
}

// Create inserts a new user
func (r *UserRepository) Create(ctx context.Context, user *models.User) error {
	query := `
		INSERT INTO users (username, password_hash, first_name, last_name, created_at, updated_at)
		VALUES ($1, $2, $3, $4, NOW(), NOW())
		RETURNING id, created_at, updated_at
	`

	err := r.DB().QueryRow(
		ctx,
		query,
		user.Username,
		user.PasswordHash,
		user.FirstName,
		user.LastName,
	).Scan(&user.ID, &user.CreatedAt, &user.UpdatedAt)

	if err != nil {
		return fmt.Errorf("failed to create user: %w", err)
	}

	return nil
}

// FindByID retrieves a user by ID
func (r *UserRepository) FindByID(ctx context.Context, id int64) (*models.User, error) {
	query := `SELECT id, username, password_hash, first_name, last_name, created_at, updated_at FROM users WHERE id = $1`

	var user models.User
	err := r.DB().QueryRow(ctx, query, id).Scan(
		&user.ID,
		&user.Username,
		&user.PasswordHash,
		&user.FirstName,
		&user.LastName,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if err == pgx.ErrNoRows {
		return nil, fmt.Errorf("user not found")
	}

	if err != nil {
		return nil, fmt.Errorf("failed to find user: %w", err)
	}

	return &user, nil
}

// FindByUsername retrieves a user by username
func (r *UserRepository) FindByUsername(ctx context.Context, username string) (*models.User, error) {
	query := `SELECT id, username, password_hash, first_name, last_name, created_at, updated_at FROM users WHERE username = $1`

	var user models.User
	err := r.DB().QueryRow(ctx, query, username).Scan(
		&user.ID,
		&user.Username,
		&user.PasswordHash,
		&user.FirstName,
		&user.LastName,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if err == pgx.ErrNoRows {
		return nil, fmt.Errorf("user not found")
	}

	if err != nil {
		return nil, fmt.Errorf("failed to find user by username: %w", err)
	}

	return &user, nil
}

// Update modifies an existing user
func (r *UserRepository) Update(ctx context.Context, user *models.User) error {
	query := `
		UPDATE users
		SET username = $1, password_hash = $2, first_name = $3, last_name = $4, updated_at = NOW()
		WHERE id = $5
		RETURNING updated_at
	`

	err := r.DB().QueryRow(
		ctx,
		query,
		user.Username,
		user.PasswordHash,
		user.FirstName,
		user.LastName,
		user.ID,
	).Scan(&user.UpdatedAt)

	if err == pgx.ErrNoRows {
		return fmt.Errorf("user not found")
	}

	if err != nil {
		return fmt.Errorf("failed to update user: %w", err)
	}

	return nil
}

// UpdatePassword updates only the user's password hash
func (r *UserRepository) UpdatePassword(ctx context.Context, userID int64, newPasswordHash string) error {
	query := `
		UPDATE users
		SET password_hash = $1, updated_at = NOW()
		WHERE id = $2
	`

	result, err := r.DB().Exec(ctx, query, newPasswordHash, userID)
	if err != nil {
		return fmt.Errorf("failed to update password: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("user not found")
	}

	return nil
}

// List retrieves users with pagination
func (r *UserRepository) List(ctx context.Context, limit, offset int) ([]*models.User, error) {
	query := `
		SELECT id, username, password_hash, first_name, last_name, created_at, updated_at
		FROM users
		ORDER BY created_at DESC
		LIMIT $1 OFFSET $2
	`

	rows, err := r.DB().Query(ctx, query, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to list users: %w", err)
	}
	defer rows.Close()

	var users []*models.User
	for rows.Next() {
		var user models.User
		err := rows.Scan(
			&user.ID,
			&user.Username,
			&user.PasswordHash,
			&user.FirstName,
			&user.LastName,
			&user.CreatedAt,
			&user.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan user: %w", err)
		}
		users = append(users, &user)
	}

	return users, nil
}

// UsernameExists checks if a username is already taken
func (r *UserRepository) UsernameExists(ctx context.Context, username string) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM users WHERE username = $1)`

	var exists bool
	err := r.DB().QueryRow(ctx, query, username).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("failed to check username existence: %w", err)
	}

	return exists, nil
}

// SearchByName searches users by first or last name
func (r *UserRepository) SearchByName(ctx context.Context, searchTerm string) ([]*models.User, error) {
	query := `
		SELECT id, username, password_hash, first_name, last_name, created_at, updated_at
		FROM users
		WHERE first_name ILIKE $1 OR last_name ILIKE $1
		ORDER BY first_name ASC, last_name ASC
	`

	rows, err := r.DB().Query(ctx, query, "%"+searchTerm+"%")
	if err != nil {
		return nil, fmt.Errorf("failed to search users: %w", err)
	}
	defer rows.Close()

	var users []*models.User
	for rows.Next() {
		var user models.User
		err := rows.Scan(
			&user.ID,
			&user.Username,
			&user.PasswordHash,
			&user.FirstName,
			&user.LastName,
			&user.CreatedAt,
			&user.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan user: %w", err)
		}
		users = append(users, &user)
	}

	return users, nil
}

// CreateBatch creates multiple users (useful for seeding/importing)
func (r *UserRepository) CreateBatch(ctx context.Context, users []*models.User) error {
	if len(users) == 0 {
		return nil
	}

	g, gCtx := errgroup.WithContext(ctx)
	sem := make(chan struct{}, 10) // Max 10 concurrent

	for _, user := range users {
		user := user // Capture

		g.Go(func() error {
			sem <- struct{}{}
			defer func() { <-sem }()

			return r.Create(gCtx, user)
		})
	}

	return g.Wait()
}

// FindByIDs retrieves multiple users by their IDs concurrently
func (r *UserRepository) FindByIDs(ctx context.Context, ids []int64) ([]*models.User, error) {
	if len(ids) == 0 {
		return []*models.User{}, nil
	}

	g, gCtx := errgroup.WithContext(ctx)
	sem := make(chan struct{}, 10)
	results := make([]*models.User, len(ids))

	for i, id := range ids {
		i, id := i, id // Capture

		g.Go(func() error {
			sem <- struct{}{}
			defer func() { <-sem }()

			user, err := r.FindByID(gCtx, id)
			if err != nil {
				return fmt.Errorf("failed to fetch user %d: %w", id, err)
			}

			results[i] = user
			return nil
		})
	}

	if err := g.Wait(); err != nil {
		return nil, err
	}

	return results, nil
}
