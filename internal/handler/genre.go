package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"golang-project/internal/models"
	"golang-project/internal/service"
)

type GenreHandler struct {
	service *service.GenreService
}

func NewGenreHandler(s *service.GenreService) *GenreHandler {
	return &GenreHandler{service: s}
}

func (h *GenreHandler) List(c *gin.Context) {
	genres, err := h.service.List(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list genres"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": genres})
}

func (h *GenreHandler) Get(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	genre, err := h.service.Get(c.Request.Context(), id)
	if err != nil {
		if err == service.ErrGenreNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "genre not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get genre"})
		return
	}
	c.JSON(http.StatusOK, genre)
}

func (h *GenreHandler) Create(c *gin.Context) {
	var req models.CreateGenreRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}
	genre, err := h.service.Create(c.Request.Context(), req)
	if err != nil {
		switch err {
		case service.ErrGenreExists:
			c.JSON(http.StatusConflict, gin.H{"error": "genre already exists"})
		default:
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		}
		return
	}
	c.JSON(http.StatusCreated, genre)
}

func (h *GenreHandler) Update(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	var req models.CreateGenreRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}
	genre, err := h.service.Update(c.Request.Context(), id, req)
	if err != nil {
		switch err {
		case service.ErrGenreNotFound:
			c.JSON(http.StatusNotFound, gin.H{"error": "genre not found"})
		case service.ErrGenreExists:
			c.JSON(http.StatusConflict, gin.H{"error": "genre already exists"})
		default:
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		}
		return
	}
	c.JSON(http.StatusOK, genre)
}

func (h *GenreHandler) Delete(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	if err := h.service.Delete(c.Request.Context(), id); err != nil {
		if err == service.ErrGenreNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "genre not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete genre"})
		return
	}
	c.Status(http.StatusNoContent)
}
