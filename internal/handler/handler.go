package handler

import (
	"database/sql"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"

	"golang-project/internal/repository"
	"golang-project/internal/service"
)

func SetupRoutes(db *sql.DB, jwtSecret string) *gin.Engine {
	router := gin.Default()

	v := validator.New()
	userRepo := repository.NewUserRepository(db)
	authService := service.NewAuthService(userRepo, v, jwtSecret)
	authHandler := NewAuthHandler(authService)

	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})
	router.POST("/auth/register", authHandler.Register)
	router.POST("/auth/login", authHandler.Login)

	return router
}
