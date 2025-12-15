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

	genreRepo := repository.NewGenreRepository(db)
	movieRepo := repository.NewMovieRepository(db)
	reviewRepo := repository.NewReviewRepository(db)
	genreService := service.NewGenreService(genreRepo, v)
	movieService := service.NewMovieService(movieRepo, genreRepo, v)
	reviewService := service.NewReviewService(reviewRepo, movieRepo, v)
	genreHandler := NewGenreHandler(genreService)
	movieHandler := NewMovieHandler(movieService)
	reviewHandler := NewReviewHandler(reviewService)

	public := router.Group("/")
	public.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})
	public.POST("/auth/register", authHandler.Register)
	public.POST("/auth/login", authHandler.Login)
	public.GET("/genres", genreHandler.List)
	public.GET("/genres/:id", genreHandler.Get)
	public.GET("/movies", movieHandler.List)
	public.GET("/movies/:id", movieHandler.Get)
	public.GET("/movies/:id/reviews", reviewHandler.ListByMovie)

	protected := router.Group("/api", middleware.AuthMiddleware(jwtSecret))
	protected.GET("/me", func(c *gin.Context) {
		userID, _ := c.Get(string(middleware.ContextUserID))
		role, _ := c.Get(string(middleware.ContextRole))
		c.JSON(http.StatusOK, gin.H{
			"user_id": userID,
			"role":    role,
		})
	})

	admin := protected.Group("/", middleware.RequireRoles("admin"))
	admin.POST("/genres", genreHandler.Create)
	admin.PUT("/genres/:id", genreHandler.Update)
	admin.DELETE("/genres/:id", genreHandler.Delete)

	admin.POST("/movies", movieHandler.Create)
	admin.PUT("/movies/:id", movieHandler.Update)
	admin.DELETE("/movies/:id", movieHandler.Delete)

	protected.POST("/movies/:id/reviews", reviewHandler.Create)
	protected.PUT("/reviews/:id", reviewHandler.Update)
	protected.DELETE("/reviews/:id", reviewHandler.Delete)

	return router
}
