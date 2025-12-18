package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"golang-project/internal/models"
	"golang-project/internal/service"
)

type MovieHandler struct {
	service *service.MovieService
}

func NewMovieHandler(s *service.MovieService) *MovieHandler {
	return &MovieHandler{service: s}
}

func (h *MovieHandler) List(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))

	var filters models.MovieFilters
	filters.Genre = c.Query("genre")
	filters.Search = c.Query("search")
	filters.Sort = c.Query("sort")
	if yearStr := c.Query("year"); yearStr != "" {
		if year, err := strconv.Atoi(yearStr); err == nil {
			filters.Year = year
		}
	}
	if minRatingStr := c.Query("min_rating"); minRatingStr != "" {
		if rating, err := strconv.ParseFloat(minRatingStr, 64); err == nil {
			filters.MinRating = rating
		}
	}
	if genreIDStr := c.Query("genre_id"); genreIDStr != "" {
		if genreID, err := strconv.Atoi(genreIDStr); err == nil {
			filters.GenreID = &genreID
		}
	}

	resp, err := h.service.List(c.Request.Context(), filters, page, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list movies"})
		return
	}
	c.JSON(http.StatusOK, resp)
}

func (h *MovieHandler) Get(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	movie, err := h.service.Get(c.Request.Context(), id)
	if err != nil {
		if err == service.ErrMovieNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "movie not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get movie"})
		return
	}
	c.JSON(http.StatusOK, movie)
}

func (h *MovieHandler) Create(c *gin.Context) {
	var req models.CreateMovieRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}
	movie, err := h.service.Create(c.Request.Context(), req)
	if err != nil {
		switch err {
		case service.ErrNoGenresProvided:
			c.JSON(http.StatusBadRequest, gin.H{"error": "genre_ids required"})
		case service.ErrGenreNotFound:
			c.JSON(http.StatusBadRequest, gin.H{"error": "genre not found"})
		default:
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		}
		return
	}
	c.JSON(http.StatusCreated, movie)
}

func (h *MovieHandler) Update(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	var req models.UpdateMovieRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	movie, err := h.service.Update(c.Request.Context(), id, req)
	if err != nil {
		switch err {
		case service.ErrMovieNotFound:
			c.JSON(http.StatusNotFound, gin.H{"error": "movie not found"})
		case service.ErrNoGenresProvided:
			c.JSON(http.StatusBadRequest, gin.H{"error": "genre_ids required"})
		case service.ErrGenreNotFound:
			c.JSON(http.StatusBadRequest, gin.H{"error": "genre not found"})
		default:
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		}
		return
	}
	c.JSON(http.StatusOK, movie)
}

func (h *MovieHandler) Delete(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	if err := h.service.Delete(c.Request.Context(), id); err != nil {
		if err == service.ErrMovieNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "movie not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete movie"})
		return
	}
	c.Status(http.StatusNoContent)
}
