package todo

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Repository provides Cockroach-backed persistence for todos.
type Repository struct {
	pool *pgxpool.Pool
}

// NewRepository constructs a repository around the supplied pgx pool.
func NewRepository(pool *pgxpool.Pool) *Repository {
	return &Repository{pool: pool}
}

// Create inserts a todo row.
func (r *Repository) Create(ctx context.Context, input CreateInput) (Todo, error) {
	query := `
		INSERT INTO todos (id, user_id, title, description, due_date)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, user_id, title, description, due_date, completed, created_at, updated_at
	`

	id := uuid.New()
	var t Todo

	if err := r.pool.QueryRow(ctx, query,
		id,
		input.UserID,
		input.Title,
		input.Description,
		input.DueDate,
	).Scan(
		&t.ID,
		&t.UserID,
		&t.Title,
		&t.Description,
		&t.DueDate,
		&t.Completed,
		&t.CreatedAt,
		&t.UpdatedAt,
	); err != nil {
		return Todo{}, fmt.Errorf("insert todo: %w", err)
	}

	return t, nil
}

// Get fetches a todo by id.
func (r *Repository) Get(ctx context.Context, id uuid.UUID) (Todo, error) {
	query := `
		SELECT id, user_id, title, description, due_date, completed, created_at, updated_at
		FROM todos
		WHERE id = $1
	`

	var t Todo
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&t.ID,
		&t.UserID,
		&t.Title,
		&t.Description,
		&t.DueDate,
		&t.Completed,
		&t.CreatedAt,
		&t.UpdatedAt,
	)

	switch {
	case err == nil:
		return t, nil
	case err == pgx.ErrNoRows:
		return Todo{}, ErrNotFound
	default:
		return Todo{}, fmt.Errorf("select todo: %w", err)
	}
}

// ListByUser returns todos scoped to a user.
func (r *Repository) ListByUser(ctx context.Context, userID uuid.UUID) ([]Todo, error) {
	query := `
		SELECT id, user_id, title, description, due_date, completed, created_at, updated_at
		FROM todos
		WHERE user_id = $1
		ORDER BY created_at DESC
	`

	rows, err := r.pool.Query(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("query todos: %w", err)
	}
	defer rows.Close()

	var result []Todo
	for rows.Next() {
		var t Todo
		if err := rows.Scan(
			&t.ID,
			&t.UserID,
			&t.Title,
			&t.Description,
			&t.DueDate,
			&t.Completed,
			&t.CreatedAt,
			&t.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan todo: %w", err)
		}
		result = append(result, t)
	}

	if rows.Err() != nil {
		return nil, fmt.Errorf("iterate todos: %w", rows.Err())
	}

	return result, nil
}

// Update applies partial updates to a todo and returns the new state.
func (r *Repository) Update(ctx context.Context, id uuid.UUID, input UpdateInput) (Todo, error) {
	setClauses := make([]string, 0, 5)
	args := make([]any, 0, 5)
	position := 1

	if input.Title != nil {
		setClauses = append(setClauses, fmt.Sprintf("title = $%d", position))
		args = append(args, *input.Title)
		position++
	}

	if input.Description != nil {
		setClauses = append(setClauses, fmt.Sprintf("description = $%d", position))
		args = append(args, *input.Description)
		position++
	}

	if input.Completed != nil {
		setClauses = append(setClauses, fmt.Sprintf("completed = $%d", position))
		args = append(args, *input.Completed)
		position++
	}

	if input.ClearDueDate {
		setClauses = append(setClauses, "due_date = NULL")
	} else if input.DueDate != nil {
		setClauses = append(setClauses, fmt.Sprintf("due_date = $%d", position))
		args = append(args, *input.DueDate)
		position++
	}

	if len(setClauses) == 0 {
		return r.Get(ctx, id)
	}

	setClauses = append(setClauses, "updated_at = current_timestamp")
	args = append(args, id)

	query := fmt.Sprintf(`
		UPDATE todos
		SET %s
		WHERE id = $%d
		RETURNING id, user_id, title, description, due_date, completed, created_at, updated_at
	`, strings.Join(setClauses, ", "), position)

	var t Todo
	if err := r.pool.QueryRow(ctx, query, args...).Scan(
		&t.ID,
		&t.UserID,
		&t.Title,
		&t.Description,
		&t.DueDate,
		&t.Completed,
		&t.CreatedAt,
		&t.UpdatedAt,
	); err != nil {
		if err == pgx.ErrNoRows {
			return Todo{}, ErrNotFound
		}
		return Todo{}, fmt.Errorf("update todo: %w", err)
	}

	return t, nil
}

// Delete removes a todo.
func (r *Repository) Delete(ctx context.Context, id uuid.UUID) error {
	tag, err := r.pool.Exec(ctx, `DELETE FROM todos WHERE id = $1`, id)
	if err != nil {
		return fmt.Errorf("delete todo: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

// ListDueWithin returns incomplete todos that are due within the provided window.
func (r *Repository) ListDueWithin(ctx context.Context, window time.Duration) ([]Todo, error) {
	target := time.Now().Add(window)

	query := `
		SELECT id, user_id, title, description, due_date, completed, created_at, updated_at
		FROM todos
		WHERE completed = FALSE
		  AND due_date IS NOT NULL
		  AND due_date <= $1
		ORDER BY due_date ASC
	`

	rows, err := r.pool.Query(ctx, query, target)
	if err != nil {
		return nil, fmt.Errorf("query due todos: %w", err)
	}
	defer rows.Close()

	var result []Todo
	for rows.Next() {
		var t Todo
		if err := rows.Scan(
			&t.ID,
			&t.UserID,
			&t.Title,
			&t.Description,
			&t.DueDate,
			&t.Completed,
			&t.CreatedAt,
			&t.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan due todo: %w", err)
		}
		result = append(result, t)
	}

	if rows.Err() != nil {
		return nil, fmt.Errorf("iterate due todos: %w", rows.Err())
	}

	return result, nil
}
