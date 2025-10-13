package todo

import (
	"time"

	"github.com/google/uuid"
)

// Todo represents a task owned by a user.
type Todo struct {
	ID          uuid.UUID  `json:"id"`
	UserID      uuid.UUID  `json:"user_id"`
	Title       string     `json:"title"`
	Description string     `json:"description,omitempty"`
	DueDate     *time.Time `json:"due_date,omitempty"`
	Completed   bool       `json:"completed"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}

// CreateInput holds the payload required to create a todo.
type CreateInput struct {
	UserID      uuid.UUID  `json:"user_id" binding:"required"`
	Title       string     `json:"title" binding:"required"`
	Description string     `json:"description"`
	DueDate     *time.Time `json:"due_date"`
}

// UpdateInput allows partial updates to a todo.
type UpdateInput struct {
	Title        *string     `json:"title"`
	Description  *string     `json:"description"`
	DueDate      *time.Time  `json:"due_date"`
	Completed    *bool       `json:"completed"`
	ClearDueDate bool        `json:"clear_due_date"`
}
