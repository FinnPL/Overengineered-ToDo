package user

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// RegisterRoutes wires the user HTTP handlers onto the supplied router group.
func RegisterRoutes(router *gin.RouterGroup, repo *Repository) {
	handler := &Handler{repo: repo}

	router.POST("", handler.createUser)
	router.GET("/:id", handler.getUser)
	router.GET("", handler.listUsers)
	router.DELETE("/:id", handler.deleteUser)
}

// Handler aggregates HTTP endpoints for the user resource.
type Handler struct {
	repo *Repository
}

func (h *Handler) createUser(c *gin.Context) {
	var input CreateUserInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	user, err := h.repo.Create(c.Request.Context(), input)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, user)
}

func (h *Handler) getUser(c *gin.Context) {
	id, err := parseUUIDParam(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	user, err := h.repo.GetByID(c.Request.Context(), id)
	switch {
	case err == nil:
		c.JSON(http.StatusOK, user)
	case err == ErrNotFound:
		c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
	default:
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
	}
}

func (h *Handler) listUsers(c *gin.Context) {
	limit := 100
	if raw := c.Query("limit"); raw != "" {
		if parsed, err := strconv.Atoi(raw); err == nil && parsed > 0 {
			limit = parsed
		}
	}

	users, err := h.repo.List(c.Request.Context(), limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, users)
}

func (h *Handler) deleteUser(c *gin.Context) {
	id, err := parseUUIDParam(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.repo.Delete(c.Request.Context(), id); err != nil {
		if err == ErrNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.Status(http.StatusNoContent)
}

func parseUUIDParam(raw string) (uuid.UUID, error) {
	return uuid.Parse(raw)
}
