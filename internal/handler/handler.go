package handler

import (
	"database/sql"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"

	"golang-project/internal/middleware"
	"golang-project/internal/repository"
	"golang-project/internal/router"
	"golang-project/internal/service"
	"golang-project/pkg/jwt"
)

type jwtPasswordHasher struct{}

func (h *jwtPasswordHasher) HashPassword(password string) (string, error) {
	return jwt.HashPassword(password)
}

func (h *jwtPasswordHasher) CheckPassword(hash, password string) error {
	return jwt.CheckPassword(hash, password)
}

func SetupRoutes(db *sql.DB, jwtSecret string, events chan service.ReviewEvent) *gin.Engine {
	router := router.New()

	v := validator.New()
	userRepo := repository.NewUserRepository(db)
	authService := service.NewAuthService(userRepo, v, jwtSecret)
	authHandler := NewAuthHandler(authService)
	reviewRepo := repository.NewReviewRepository(db)
	passwordHasher := &jwtPasswordHasher{}
	userService := service.NewUserService(userRepo, reviewRepo, v, passwordHasher)

	genreRepo := repository.NewGenreRepository(db)
	movieRepo := repository.NewMovieRepository(db)
	genreService := service.NewGenreService(genreRepo, v)
	movieService := service.NewMovieService(movieRepo, genreRepo, v)
	reviewService := service.NewReviewService(reviewRepo, movieRepo, v, events)
	genreHandler := NewGenreHandler(genreService)
	movieHandler := NewMovieHandler(movieService)
	reviewHandler := NewReviewHandler(reviewService)
	auditRepo := repository.NewAuditRepository(db)
	userHandler := NewUserHandler(userService, reviewService, userRepo, movieRepo, reviewRepo, genreRepo, auditRepo)

	api := router.Group("/api/v1")

	public := api.Group("/")
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

	protected := api.Group("/", middleware.AuthMiddleware(jwtSecret))
	protected.GET("/me", userHandler.Me)
	protected.PUT("/me", userHandler.UpdateProfile)
	protected.PUT("/me/password", userHandler.UpdatePassword)
	protected.GET("/me/reviews", userHandler.MyReviews)
	protected.POST("/movies/:id/reviews", reviewHandler.Create)
	protected.PUT("/reviews/:id", reviewHandler.Update)
	protected.DELETE("/reviews/:id", reviewHandler.Delete)

	api.GET("/users/:id/reviews", userHandler.UserReviews)

	admin := api.Group("/", middleware.AuthMiddleware(jwtSecret), middleware.RequireRoles("admin"))
	admin.GET("/users", userHandler.ListUsers)
	admin.GET("/users/:id", userHandler.GetUser)
	admin.PUT("/users/:id", userHandler.UpdateUser)
	admin.PUT("/users/:id/role", userHandler.UpdateRole)
	admin.DELETE("/users/:id", userHandler.DeleteUser)
	admin.GET("/stats", userHandler.GetStats)
	admin.GET("/audit-logs", userHandler.ListAuditLogs)
	admin.POST("/genres", genreHandler.Create)
	admin.PUT("/genres/:id", genreHandler.Update)
	admin.DELETE("/genres/:id", genreHandler.Delete)

	admin.POST("/movies", movieHandler.Create)
	admin.PUT("/movies/:id", movieHandler.Update)
	admin.DELETE("/movies/:id", movieHandler.Delete)

	return router
}
