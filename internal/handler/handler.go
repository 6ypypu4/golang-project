package handler

import (
	"golang-project/internal/repository"

	"github.com/gin-gonic/gin"
)

func SetupRoutes(repos *repository.Repositories) *gin.Engine {
	router := gin.Default()

	// TODO: Setup routes with handlers that use repos
	// Example:
	// userHandler := NewUserHandler(repos.User)
	// router.POST("/users", userHandler.Create)
	// router.GET("/users/:id", userHandler.GetByID)

	return router
}
