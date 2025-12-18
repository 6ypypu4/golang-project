package handler

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"

	"golang-project/internal/middleware"
	"golang-project/internal/models"
	"golang-project/internal/repository"
	"golang-project/internal/service"
)

type UserHandler struct {
	users      *service.UserService
	reviews    *service.ReviewService
	userRepo   repository.UserRepository
	movieRepo  service.MovieCountRepo
	reviewRepo service.ReviewCountRepo
	genreRepo  service.GenreCountRepo
	auditRepo  service.AuditLogRepo
}

func NewUserHandler(users *service.UserService, reviews *service.ReviewService, userRepo repository.UserRepository, movieRepo service.MovieCountRepo, reviewRepo service.ReviewCountRepo, genreRepo service.GenreCountRepo, auditRepo service.AuditLogRepo) *UserHandler {
	return &UserHandler{
		users:      users,
		reviews:    reviews,
		userRepo:   userRepo,
		movieRepo:  movieRepo,
		reviewRepo: reviewRepo,
		genreRepo:  genreRepo,
		auditRepo:  auditRepo,
	}
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
	stats, _ := h.users.GetUserStats(c.Request.Context(), uid)

	response := gin.H{
		"user":          user,
		"reviews_count": reviewCount,
	}
	if stats != nil {
		response["average_rating"] = stats.AverageRating
		if stats.FavoriteGenre != nil {
			response["favorite_genre"] = stats.FavoriteGenre
		}
	}

	c.JSON(http.StatusOK, response)
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

type updateRoleRequest struct {
	Role string `json:"role" validate:"required,oneof=user admin"`
}

func (h *UserHandler) ListUsers(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	
	filters := models.UserFilters{
		Search: c.Query("search"),
		Role:   c.Query("role"),
	}
	
	resp, err := h.users.List(c.Request.Context(), filters, page, limit)
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

func (h *UserHandler) UpdateUser(c *gin.Context) {
	uid, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user id"})
		return
	}

	var req models.UpdateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	if err := h.users.Update(c.Request.Context(), uid, req); err != nil {
		switch err {
		case service.ErrUserNotFound:
			c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		case service.ErrUserExists:
			c.JSON(http.StatusConflict, gin.H{"error": "email or username already exists"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update user"})
		}
		return
	}

	user, err := h.users.GetByID(c.Request.Context(), uid)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get updated user"})
		return
	}

	c.JSON(http.StatusOK, user)
}

func (h *UserHandler) DeleteUser(c *gin.Context) {
	uid, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user id"})
		return
	}

	adminIDStr, _ := c.Get(string(middleware.ContextUserID))
	adminID, err := strconv.Atoi(adminIDStr.(string))
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid admin user"})
		return
	}

	if err := h.users.Delete(c.Request.Context(), uid, adminID); err != nil {
		switch err {
		case service.ErrUserNotFound:
			c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		case service.ErrCannotDeleteSelf:
			c.JSON(http.StatusBadRequest, gin.H{"error": "cannot delete yourself"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete user"})
		}
		return
	}

	c.Status(http.StatusNoContent)
}

func (h *UserHandler) UpdateProfile(c *gin.Context) {
	userIDStr, _ := c.Get(string(middleware.ContextUserID))
	uid, err := strconv.Atoi(userIDStr.(string))
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid user"})
		return
	}

	var req models.UpdateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	user, err := h.users.UpdateProfile(c.Request.Context(), uid, req)
	if err != nil {
		switch err {
		case service.ErrUserNotFound:
			c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		case service.ErrUserExists:
			c.JSON(http.StatusConflict, gin.H{"error": "email or username already exists"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update profile"})
		}
		return
	}

	c.JSON(http.StatusOK, user)
}

func (h *UserHandler) UpdatePassword(c *gin.Context) {
	userIDStr, _ := c.Get(string(middleware.ContextUserID))
	uid, err := strconv.Atoi(userIDStr.(string))
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid user"})
		return
	}

	var req models.UpdatePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	if err := h.users.UpdatePassword(c.Request.Context(), uid, req.CurrentPassword, req.NewPassword); err != nil {
		switch err {
		case service.ErrUserNotFound:
			c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		case service.ErrInvalidCredentials:
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid current password"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update password"})
		}
		return
	}

	c.Status(http.StatusNoContent)
}

func (h *UserHandler) GetStats(c *gin.Context) {
	stats, err := h.users.GetAdminStats(
		c.Request.Context(),
		h.userRepo,
		h.movieRepo,
		h.reviewRepo,
		h.genreRepo,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get stats"})
		return
	}

	c.JSON(http.StatusOK, stats)
}

func (h *UserHandler) ListAuditLogs(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))

	filters := models.AuditLogFilters{
		Event: c.Query("event"),
	}

	if userIDStr := c.Query("user_id"); userIDStr != "" {
		if userID, err := strconv.Atoi(userIDStr); err == nil {
			filters.UserID = &userID
		}
	}

	if fromDateStr := c.Query("from_date"); fromDateStr != "" {
		if fromDate, err := time.Parse("2006-01-02", fromDateStr); err == nil {
			filters.FromDate = &fromDate
		}
	}

	if toDateStr := c.Query("to_date"); toDateStr != "" {
		if toDate, err := time.Parse("2006-01-02", toDateStr); err == nil {
			filters.ToDate = &toDate
		}
	}

	resp, err := h.users.ListAuditLogs(c.Request.Context(), h.auditRepo, filters, page, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list audit logs"})
		return
	}

	c.JSON(http.StatusOK, resp)
}
