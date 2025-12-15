package handler

import (
	"database/sql"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"

	"golang-project/internal/middleware"
	"golang-project/internal/repository"
	"golang-project/internal/service"
)

func SetupRoutes(db *sql.DB, jwtSecret string) *gin.Engine {
	router := gin.New()
	router.Use(gin.Logger(), gin.Recovery(), middleware.RequestID(), middleware.CORS())

	v := validator.New()
	userRepo := repository.NewUserRepository(db)
	authService := service.NewAuthService(userRepo, v, jwtSecret)
	authHandler := NewAuthHandler(authService)

	public := router.Group("/")
	public.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})
	public.POST("/auth/register", authHandler.Register)
	public.POST("/auth/login", authHandler.Login)

	protected := router.Group("/api", middleware.AuthMiddleware(jwtSecret))
	protected.GET("/me", func(c *gin.Context) {
		userID, _ := c.Get(string(middleware.ContextUserID))
		role, _ := c.Get(string(middleware.ContextRole))
		c.JSON(http.StatusOK, gin.H{
			"user_id": userID,
			"role":    role,
		})
	})

	return router
}
