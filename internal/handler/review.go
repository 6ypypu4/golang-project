package handler

import (
	"log"
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

	filters := parseReviewFilters(c)
	reviews, err := h.service.ListByMovie(c.Request.Context(), movieID, filters, page, limit)
	if err != nil {
		log.Printf("ListByMovie error: %v", err)
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
	f.Sort = c.Query("sort")
	return f
}

func (h *ReviewHandler) Create(c *gin.Context) {
	movieIDStr := c.Param("id")
	movieID, err := strconv.Atoi(movieIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid movie id"})
		return
	}

	val, ok := c.Get(string(middleware.ContextUserID))
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	userIDStr, ok := val.(string)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid user"})
		return
	}
	userID, err := strconv.Atoi(userIDStr)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid user"})
		return
	}

	var req models.CreateReviewRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	if req.Rating < 1 || req.Rating > 10 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "rating must be between 1 and 10"})
		return
	}

	review, err := h.service.Create(c.Request.Context(), movieID, userID, req)
	if err != nil {
		log.Printf("CreateReview error: %v", err)
		switch err {
		case service.ErrMovieNotFound:
			c.JSON(http.StatusNotFound, gin.H{"error": "movie not found"})
		case service.ErrReviewExists:
			c.JSON(http.StatusConflict, gin.H{"error": "review already exists"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
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

	val, ok := c.Get(string(middleware.ContextUserID))
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	userIDStr, ok := val.(string)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid user"})
		return
	}
	userID, err := strconv.Atoi(userIDStr)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid user"})
		return
	}

	var req models.UpdateReviewRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	if req.Rating != 0 && (req.Rating < 1 || req.Rating > 10) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "rating must be between 1 and 10"})
		return
	}

	review, err := h.service.Update(c.Request.Context(), reviewID, userID, req)
	if err != nil {
		log.Printf("UpdateReview error: %v", err)
		switch err {
		case service.ErrReviewNotFound:
			c.JSON(http.StatusNotFound, gin.H{"error": "review not found"})
		case service.ErrInvalidCredentials:
			c.JSON(http.StatusForbidden, gin.H{"error": "forbidden"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
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

	val, ok := c.Get(string(middleware.ContextUserID))
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	userIDStr, ok := val.(string)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid user"})
		return
	}
	userID, err := strconv.Atoi(userIDStr)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid user"})
		return
	}

	roleVal, _ := c.Get(string(middleware.ContextRole))
	isAdmin := roleVal == "admin"

	if err := h.service.Delete(c.Request.Context(), reviewID, userID, isAdmin); err != nil {
		log.Printf("DeleteReview error: %v", err)
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
