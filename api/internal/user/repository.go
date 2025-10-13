package user

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Repository provides database persistence for users.
type Repository struct {
	pool *pgxpool.Pool
}

// NewRepository constructs a Repository backed by the supplied pgx pool.
func NewRepository(pool *pgxpool.Pool) *Repository {
	return &Repository{pool: pool}
}

// Create persists a new user row.
func (r *Repository) Create(ctx context.Context, input CreateUserInput) (User, error) {
	query := `
		INSERT INTO users (id, name, email)
		VALUES ($1, $2, $3)
		RETURNING id, name, email, created_at
	`

	id := uuid.New()
	var u User

	if err := r.pool.QueryRow(ctx, query, id, input.Name, input.Email).Scan(
		&u.ID, &u.Name, &u.Email, &u.CreatedAt,
	); err != nil {
		return User{}, fmt.Errorf("insert user: %w", err)
	}

	return u, nil
}

// GetByID fetches a user by primary key.
func (r *Repository) GetByID(ctx context.Context, id uuid.UUID) (User, error) {
	query := `
		SELECT id, name, email, created_at
		FROM users
		WHERE id = $1
	`

	var u User
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&u.ID, &u.Name, &u.Email, &u.CreatedAt,
	)

	switch {
	case err == nil:
		return u, nil
	case err == pgx.ErrNoRows:
		return User{}, ErrNotFound
	default:
		return User{}, fmt.Errorf("select user: %w", err)
	}
}

// List returns all users up to the supplied limit.
func (r *Repository) List(ctx context.Context, limit int) ([]User, error) {
	query := `
		SELECT id, name, email, created_at
		FROM users
		ORDER BY created_at DESC
		LIMIT $1
	`

	rows, err := r.pool.Query(ctx, query, limit)
	if err != nil {
		return nil, fmt.Errorf("query users: %w", err)
	}
	defer rows.Close()

	var result []User
	for rows.Next() {
		var u User
		if err := rows.Scan(&u.ID, &u.Name, &u.Email, &u.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan user: %w", err)
		}
		result = append(result, u)
	}

	if rows.Err() != nil {
		return nil, fmt.Errorf("iterate users: %w", rows.Err())
	}

	return result, nil
}

// Delete removes a user record. Returns ErrNotFound when the row is absent.
func (r *Repository) Delete(ctx context.Context, id uuid.UUID) error {
	tag, err := r.pool.Exec(ctx, `DELETE FROM users WHERE id = $1`, id)
	if err != nil {
		return fmt.Errorf("delete user: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

// Touch updates the updated_at column for the user. Useful for activity tracking.
func (r *Repository) Touch(ctx context.Context, id uuid.UUID) error {
	_, err := r.pool.Exec(ctx, `UPDATE users SET updated_at = $2 WHERE id = $1`, id, time.Now().UTC())
	if err != nil {
		return fmt.Errorf("touch user: %w", err)
	}
	return nil
}
