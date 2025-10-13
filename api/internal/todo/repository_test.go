package todo

import (
	"context"
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
	input := CreateInput{
		UserID:      uuid.New(),
		Title:       "Title",
		Description: "Desc",
	}
	returnedID := uuid.New()
	now := time.Now()

	rows := pgxmock.NewRows([]string{
		"id", "user_id", "title", "description", "due_date", "completed", "created_at", "updated_at",
	}).AddRow(returnedID, input.UserID, input.Title, input.Description, nil, false, now, now)

	mock.ExpectQuery("INSERT INTO todos").
		WithArgs(pgxmock.AnyArg(), input.UserID, input.Title, input.Description, input.DueDate).
		WillReturnRows(rows)

	todo, err := repo.Create(context.Background(), input)
	require.NoError(t, err)
	require.Equal(t, returnedID, todo.ID)
	require.Equal(t, input.UserID, todo.UserID)
	require.Equal(t, input.Title, todo.Title)
	require.Equal(t, input.Description, todo.Description)
	require.Nil(t, todo.DueDate)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestRepositoryGet(t *testing.T) {
	mock, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mock.Close()

	repo := NewRepository(mock)
	id := uuid.New()
	userID := uuid.New()
	now := time.Now()

	rows := pgxmock.NewRows([]string{
		"id", "user_id", "title", "description", "due_date", "completed", "created_at", "updated_at",
	}).AddRow(id, userID, "Title", "Desc", nil, false, now, now)

	mock.ExpectQuery("SELECT id, user_id, title, description, due_date, completed, created_at, updated_at").
		WithArgs(id).
		WillReturnRows(rows)

	tt, err := repo.Get(context.Background(), id)
	require.NoError(t, err)
	require.Equal(t, id, tt.ID)
	require.Equal(t, userID, tt.UserID)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestRepositoryGetNotFound(t *testing.T) {
	mock, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mock.Close()

	repo := NewRepository(mock)
	id := uuid.New()

	mock.ExpectQuery("SELECT id, user_id, title, description, due_date, completed, created_at, updated_at").
		WithArgs(id).
		WillReturnError(pgx.ErrNoRows)

	_, err = repo.Get(context.Background(), id)
	require.ErrorIs(t, err, ErrNotFound)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestRepositoryListByUser(t *testing.T) {
	mock, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mock.Close()

	repo := NewRepository(mock)
	userID := uuid.New()
	now := time.Now()

	rows := pgxmock.NewRows([]string{
		"id", "user_id", "title", "description", "due_date", "completed", "created_at", "updated_at",
	}).AddRow(uuid.New(), userID, "A", "desc", nil, false, now, now).
		AddRow(uuid.New(), userID, "B", "desc", nil, true, now.Add(-time.Hour), now.Add(-time.Hour))

	mock.ExpectQuery("SELECT id, user_id, title, description, due_date, completed, created_at, updated_at FROM todos").
		WithArgs(userID).
		WillReturnRows(rows)

	list, err := repo.ListByUser(context.Background(), userID)
	require.NoError(t, err)
	require.Len(t, list, 2)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestRepositoryUpdate(t *testing.T) {
	mock, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mock.Close()

	repo := NewRepository(mock)
	id := uuid.New()
	title := "New Title"
	completed := true
	now := time.Now()

	rows := pgxmock.NewRows([]string{
		"id", "user_id", "title", "description", "due_date", "completed", "created_at", "updated_at",
	}).AddRow(id, uuid.New(), title, "Desc", nil, completed, now, now)

	mock.ExpectQuery("UPDATE todos SET").
		WithArgs(title, completed, id).
		WillReturnRows(rows)

	updated, err := repo.Update(context.Background(), id, UpdateInput{
		Title:     &title,
		Completed: &completed,
	})
	require.NoError(t, err)
	require.Equal(t, title, updated.Title)
	require.Equal(t, completed, updated.Completed)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestRepositoryUpdateClearDueDate(t *testing.T) {
	mock, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mock.Close()

	repo := NewRepository(mock)
	id := uuid.New()
	desc := "Updated description"
	now := time.Now()

	rows := pgxmock.NewRows([]string{
		"id", "user_id", "title", "description", "due_date", "completed", "created_at", "updated_at",
	}).AddRow(id, uuid.New(), "Title", desc, nil, false, now, now)

	mock.ExpectQuery("UPDATE todos SET").
		WithArgs(desc, id).
		WillReturnRows(rows)

	updated, err := repo.Update(context.Background(), id, UpdateInput{
		Description:  &desc,
		ClearDueDate: true,
	})
	require.NoError(t, err)
	require.Nil(t, updated.DueDate)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestRepositoryUpdateNoFieldsFallsBackToGet(t *testing.T) {
	mock, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mock.Close()

	repo := NewRepository(mock)
	id := uuid.New()
	now := time.Now()

	rows := pgxmock.NewRows([]string{
		"id", "user_id", "title", "description", "due_date", "completed", "created_at", "updated_at",
	}).AddRow(id, uuid.New(), "Title", "Desc", nil, false, now, now)

	mock.ExpectQuery("SELECT id, user_id, title, description, due_date, completed, created_at, updated_at").
		WithArgs(id).
		WillReturnRows(rows)

	updated, err := repo.Update(context.Background(), id, UpdateInput{})
	require.NoError(t, err)
	require.Equal(t, "Title", updated.Title)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestRepositoryDelete(t *testing.T) {
	mock, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mock.Close()

	repo := NewRepository(mock)
	id := uuid.New()

	mock.ExpectExec("DELETE FROM todos").
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

	mock.ExpectExec("DELETE FROM todos").
		WithArgs(id).
		WillReturnResult(pgxmock.NewResult("DELETE", 0))

	err = repo.Delete(context.Background(), id)
	require.ErrorIs(t, err, ErrNotFound)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestRepositoryListDueWithin(t *testing.T) {
	mock, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mock.Close()

	repo := NewRepository(mock)
	now := time.Now()

	rows := pgxmock.NewRows([]string{
		"id", "user_id", "title", "description", "due_date", "completed", "created_at", "updated_at",
	})
	due := now.Add(30 * time.Minute)
	rows.AddRow(uuid.New(), uuid.New(), "Due soon", "desc", &due, false, now, now)

	mock.ExpectQuery("SELECT id, user_id, title, description, due_date, completed, created_at, updated_at FROM todos WHERE completed = FALSE").
		WithArgs(pgxmock.AnyArg()).
		WillReturnRows(rows)

	result, err := repo.ListDueWithin(context.Background(), time.Hour)
	require.NoError(t, err)
	require.Len(t, result, 1)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestRepositoryListDueWithinError(t *testing.T) {
	mock, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mock.Close()

	repo := NewRepository(mock)

	mock.ExpectQuery("SELECT id, user_id, title, description, due_date, completed, created_at, updated_at FROM todos WHERE completed = FALSE").
		WithArgs(pgxmock.AnyArg()).
		WillReturnError(pgx.ErrNoRows)

	_, err = repo.ListDueWithin(context.Background(), time.Hour)
	require.Error(t, err)
	require.NoError(t, mock.ExpectationsWereMet())
}
