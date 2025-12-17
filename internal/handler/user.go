package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"golang-project/internal/middleware"
	"golang-project/internal/models"
	"golang-project/internal/service"
)

type UserHandler struct {
	users   *service.UserService
	reviews *service.ReviewService
}

func NewUserHandler(users *service.UserService, reviews *service.ReviewService) *UserHandler {
	return &UserHandler{users: users, reviews: reviews}
}

func (h *UserHandler) Me(c *gin.Context) {
	userIDStr, _ := c.Get(string(middleware.ContextUserID))
	uid, err := strconv.Atoi(userIDStr.(string))
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid user"})
		return
	}

	user, err := h.users.GetByID(c.Request.Context(), uid)
	if err != nil {
		if err == service.ErrUserNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get user"})
		return
	}

	reviewCount, _ := h.reviews.CountByUser(c.Request.Context(), uid)

	c.JSON(http.StatusOK, gin.H{
		"user":          user,
		"reviews_count": reviewCount,
	})
}

func (h *UserHandler) MyReviews(c *gin.Context) {
	userIDStr, _ := c.Get(string(middleware.ContextUserID))
	uid, err := strconv.Atoi(userIDStr.(string))
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid user"})
		return
	}
	h.listReviewsByUser(c, uid)
}

func (h *UserHandler) UserReviews(c *gin.Context) {
	uid, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user id"})
		return
	}
	h.listReviewsByUser(c, uid)
}

func (h *UserHandler) listReviewsByUser(c *gin.Context, uid int) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	filters := parseReviewFilters(c)

	reviews, err := h.reviews.ListByUser(c.Request.Context(), uid, filters, page, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list reviews"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": reviews})
}

func parseReviewFilters(c *gin.Context) models.ReviewFilters {
	f := models.ReviewFilters{}
	if minStr := c.Query("min_rating"); minStr != "" {
		if v, err := strconv.Atoi(minStr); err == nil {
			f.MinRating = v
		}
	}
	if maxStr := c.Query("max_rating"); maxStr != "" {
		if v, err := strconv.Atoi(maxStr); err == nil {
			f.MaxRating = v
		}
	}
	return f
}

type updateRoleRequest struct {
	Role string `json:"role" validate:"required,oneof=user admin"`
}

func (h *UserHandler) ListUsers(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	resp, err := h.users.List(c.Request.Context(), page, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list users"})
		return
	}
	c.JSON(http.StatusOK, resp)
}

func (h *UserHandler) GetUser(c *gin.Context) {
	uid, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user id"})
		return
	}
	user, err := h.users.GetByID(c.Request.Context(), uid)
	if err != nil {
		if err == service.ErrUserNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get user"})
		return
	}
	c.JSON(http.StatusOK, user)
}

func (h *UserHandler) UpdateRole(c *gin.Context) {
	uid, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user id"})
		return
	}
	var req updateRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}
	if err := h.users.UpdateRole(c.Request.Context(), uid, req.Role); err != nil {
		switch err {
		case service.ErrInvalidRole:
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid role"})
		case service.ErrUserNotFound:
			c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update role"})
		}
		return
	}
	c.Status(http.StatusNoContent)
}
