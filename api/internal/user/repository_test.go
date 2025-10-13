package user

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	pgxmock "github.com/pashagolub/pgxmock/v3"
	"github.com/stretchr/testify/require"
)

func TestRepositoryCreate(t *testing.T) {
	mock, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mock.Close()

	repo := NewRepository(mock)
	input := CreateUserInput{Name: "Alice", Email: "alice@example.com"}

	returnedID := uuid.New()
	createdAt := time.Now()
	rows := pgxmock.NewRows([]string{"id", "name", "email", "created_at"}).
		AddRow(returnedID, input.Name, input.Email, createdAt)

	mock.ExpectQuery("INSERT INTO users").
		WithArgs(pgxmock.AnyArg(), input.Name, input.Email).
		WillReturnRows(rows)

	u, err := repo.Create(context.Background(), input)
	require.NoError(t, err)
	require.Equal(t, returnedID, u.ID)
	require.Equal(t, input.Name, u.Name)
	require.Equal(t, input.Email, u.Email)
	require.WithinDuration(t, createdAt, u.CreatedAt, time.Millisecond)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestRepositoryGetByID(t *testing.T) {
	mock, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mock.Close()

	repo := NewRepository(mock)
	id := uuid.New()
	now := time.Now()

	rows := pgxmock.NewRows([]string{"id", "name", "email", "created_at"}).
		AddRow(id, "Bob", "bob@example.com", now)

	mock.ExpectQuery("SELECT id, name, email, created_at FROM users").
		WithArgs(id).
		WillReturnRows(rows)

	u, err := repo.GetByID(context.Background(), id)
	require.NoError(t, err)
	require.Equal(t, id, u.ID)
	require.Equal(t, "Bob", u.Name)
	require.Equal(t, "bob@example.com", u.Email)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestRepositoryGetByIDNotFound(t *testing.T) {
	mock, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mock.Close()

	repo := NewRepository(mock)
	id := uuid.New()

	mock.ExpectQuery("SELECT id, name, email, created_at FROM users").
		WithArgs(id).
		WillReturnError(pgx.ErrNoRows)

	_, err = repo.GetByID(context.Background(), id)
	require.ErrorIs(t, err, ErrNotFound)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestRepositoryList(t *testing.T) {
	mock, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mock.Close()

	repo := NewRepository(mock)
	now := time.Now()
	rows := pgxmock.NewRows([]string{"id", "name", "email", "created_at"}).
		AddRow(uuid.New(), "Alice", "alice@example.com", now).
		AddRow(uuid.New(), "Bob", "bob@example.com", now.Add(-time.Hour))

	mock.ExpectQuery("SELECT id, name, email, created_at FROM users").
		WithArgs(5).
		WillReturnRows(rows)

	users, err := repo.List(context.Background(), 5)
	require.NoError(t, err)
	require.Len(t, users, 2)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestRepositoryDelete(t *testing.T) {
	mock, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mock.Close()

	repo := NewRepository(mock)
	id := uuid.New()

	mock.ExpectExec("DELETE FROM users").
		WithArgs(id).
		WillReturnResult(pgxmock.NewResult("DELETE", 1))

	err = repo.Delete(context.Background(), id)
	require.NoError(t, err)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestRepositoryDeleteNotFound(t *testing.T) {
	mock, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mock.Close()

	repo := NewRepository(mock)
	id := uuid.New()

	mock.ExpectExec("DELETE FROM users").
		WithArgs(id).
		WillReturnResult(pgxmock.NewResult("DELETE", 0))

	err = repo.Delete(context.Background(), id)
	require.ErrorIs(t, err, ErrNotFound)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestRepositoryTouch(t *testing.T) {
	mock, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mock.Close()

	repo := NewRepository(mock)
	id := uuid.New()

	mock.ExpectExec("UPDATE users SET updated_at =").
		WithArgs(id, pgxmock.AnyArg()).
		WillReturnResult(pgxmock.NewResult("UPDATE", 1))

	err = repo.Touch(context.Background(), id)
	require.NoError(t, err)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestRepositoryCreateErrorPropagates(t *testing.T) {
	mock, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mock.Close()

	repo := NewRepository(mock)
	input := CreateUserInput{Name: "Alice", Email: "alice@example.com"}

	mock.ExpectQuery("INSERT INTO users").
		WithArgs(pgxmock.AnyArg(), input.Name, input.Email).
		WillReturnError(errors.New("boom"))

	_, err = repo.Create(context.Background(), input)
	require.Error(t, err)
	require.NoError(t, mock.ExpectationsWereMet())
}
