package todo

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// RegisterRoutes wires the todo HTTP handlers to a sub-router.
func RegisterRoutes(router *gin.RouterGroup, repo *Repository) {
	handler := &Handler{repo: repo}

	router.POST("", handler.createTodo)
	router.GET("/:id", handler.getTodo)
	router.GET("", handler.listTodos)
	router.PUT("/:id", handler.updateTodo)
	router.DELETE("/:id", handler.deleteTodo)
	router.PATCH("/:id/complete", handler.markComplete)
}

// Handler exposes HTTP endpoints for todos.
type Handler struct {
	repo *Repository
}

func (h *Handler) createTodo(c *gin.Context) {
	var input CreateInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if input.UserID == uuid.Nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "user_id is required"})
		return
	}

	t, err := h.repo.Create(c.Request.Context(), input)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, t)
}

func (h *Handler) getTodo(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	t, err := h.repo.Get(c.Request.Context(), id)
	switch {
	case err == nil:
		c.JSON(http.StatusOK, t)
	case err == ErrNotFound:
		c.JSON(http.StatusNotFound, gin.H{"error": "todo not found"})
	default:
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
	}
}

func (h *Handler) listTodos(c *gin.Context) {
	userIDParam := c.Query("user_id")
	if userIDParam == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "user_id query parameter is required"})
		return
	}

	userID, err := uuid.Parse(userIDParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user_id"})
		return
	}

	todos, err := h.repo.ListByUser(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, todos)
}

func (h *Handler) updateTodo(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var input UpdateInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if input.ClearDueDate && input.DueDate != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "due_date and clear_due_date are mutually exclusive"})
		return
	}

	t, err := h.repo.Update(c.Request.Context(), id, input)
	switch {
	case err == nil:
		c.JSON(http.StatusOK, t)
	case err == ErrNotFound:
		c.JSON(http.StatusNotFound, gin.H{"error": "todo not found"})
	default:
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
	}
}

func (h *Handler) deleteTodo(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.repo.Delete(c.Request.Context(), id); err != nil {
		if err == ErrNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "todo not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.Status(http.StatusNoContent)
}

func (h *Handler) markComplete(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	input := UpdateInput{
		Completed: ptrTo(true),
	}

	t, err := h.repo.Update(c.Request.Context(), id, input)
	switch {
	case err == nil:
		c.JSON(http.StatusOK, t)
	case err == ErrNotFound:
		c.JSON(http.StatusNotFound, gin.H{"error": "todo not found"})
	default:
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
	}
}

func ptrTo[T any](v T) *T {
	return &v
}
