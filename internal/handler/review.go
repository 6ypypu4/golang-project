package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"golang-project/internal/middleware"
	"golang-project/internal/models"
	"golang-project/internal/service"
)

type ReviewHandler struct {
	service *service.ReviewService
}

func NewReviewHandler(s *service.ReviewService) *ReviewHandler {
	return &ReviewHandler{service: s}
}

func (h *ReviewHandler) ListByMovie(c *gin.Context) {
	movieIDStr := c.Param("id")
	movieID, err := strconv.Atoi(movieIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid movie id"})
		return
	}
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))

	reviews, err := h.service.ListByMovie(c.Request.Context(), movieID, page, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list reviews"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": reviews})
}

func (h *ReviewHandler) Create(c *gin.Context) {
	movieIDStr := c.Param("id")
	movieID, err := strconv.Atoi(movieIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid movie id"})
		return
	}
	userIDStr, _ := c.Get(string(middleware.ContextUserID))
	userID, err := strconv.Atoi(userIDStr.(string))
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid user"})
		return
	}

	var req models.CreateReviewRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	review, err := h.service.Create(c.Request.Context(), movieID, userID, req)
	if err != nil {
		switch err {
		case service.ErrMovieNotFound:
			c.JSON(http.StatusNotFound, gin.H{"error": "movie not found"})
		case service.ErrReviewExists:
			c.JSON(http.StatusConflict, gin.H{"error": "review already exists"})
		default:
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		}
		return
	}
	c.JSON(http.StatusCreated, review)
}

func (h *ReviewHandler) Update(c *gin.Context) {
	reviewIDStr := c.Param("id")
	reviewID, err := strconv.Atoi(reviewIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid review id"})
		return
	}
	userIDStr, _ := c.Get(string(middleware.ContextUserID))
	userID, err := strconv.Atoi(userIDStr.(string))
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid user"})
		return
	}

	var req models.UpdateReviewRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	review, err := h.service.Update(c.Request.Context(), reviewID, userID, req)
	if err != nil {
		switch err {
		case service.ErrReviewNotFound:
			c.JSON(http.StatusNotFound, gin.H{"error": "review not found"})
		case service.ErrInvalidCredentials:
			c.JSON(http.StatusForbidden, gin.H{"error": "forbidden"})
		default:
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		}
		return
	}
	c.JSON(http.StatusOK, review)
}

func (h *ReviewHandler) Delete(c *gin.Context) {
	reviewIDStr := c.Param("id")
	reviewID, err := strconv.Atoi(reviewIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid review id"})
		return
	}
	userIDStr, _ := c.Get(string(middleware.ContextUserID))
	userID, err := strconv.Atoi(userIDStr.(string))
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid user"})
		return
	}
	roleVal, _ := c.Get(string(middleware.ContextRole))
	isAdmin := roleVal == "admin"

	if err := h.service.Delete(c.Request.Context(), reviewID, userID, isAdmin); err != nil {
		switch err {
		case service.ErrReviewNotFound:
			c.JSON(http.StatusNotFound, gin.H{"error": "review not found"})
		case service.ErrInvalidCredentials:
			c.JSON(http.StatusForbidden, gin.H{"error": "forbidden"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete review"})
		}
		return
	}
	c.Status(http.StatusNoContent)
}
