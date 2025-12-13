package router

import (
	"golang-project/internal/handler"
	"golang-project/internal/middleware"
	"golang-project/pkg/jwt"

	"github.com/gin-gonic/gin"
)

type Handlers struct {
	Auth   *handler.AuthHandler
	Movie  *handler.MovieHandler
	Genre  *handler.GenreHandler
	Review *handler.ReviewHandler
}

func New(h Handlers,jwtMgr *jwt.Manager,) *gin.Engine {

	r := gin.Default()
	r.Use(middleware.CORSMiddleware())

	api := r.Group("/api/v1")

	auth := api.Group("/auth")
	auth.POST("/register", h.Auth.Register)
	auth.POST("/login", h.Auth.Login)

	protected := api.Group("")
	protected.Use(middleware.AuthMiddleware(jwtMgr))
	protected.GET("/users/me", h.Auth.GetProfile)
	protected.GET("/users/me/reviews", h.Review.GetByUserID)

	movies := api.Group("/movies")
	movies.GET("", h.Movie.GetAll)
	movies.GET("/:id", h.Movie.GetByID)
	movies.GET("/:id/reviews", h.Review.GetByMovieID)

	adminMovies := movies.Group("")
	adminMovies.Use(
		middleware.AuthMiddleware(jwtMgr),
		middleware.AdminMiddleware(),
	)
	adminMovies.POST("", h.Movie.Create)
	adminMovies.PUT("/:id", h.Movie.Update)
	adminMovies.DELETE("/:id", h.Movie.Delete)

	genres := api.Group("/genres")
	genres.GET("", h.Genre.GetAll)
	genres.GET("/:id", h.Genre.GetByID)

	adminGenres := genres.Group("")
	adminGenres.Use(
		middleware.AuthMiddleware(jwtMgr),
		middleware.AdminMiddleware(),
	)
	adminGenres.POST("", h.Genre.Create)
	adminGenres.PUT("/:id", h.Genre.Update)
	adminGenres.DELETE("/:id", h.Genre.Delete)

	reviews := api.Group("/reviews")
	reviews.Use(middleware.AuthMiddleware(jwtMgr))
	reviews.GET("/:id", h.Review.GetByID)
	reviews.PUT("/:id", h.Review.Update)
	reviews.DELETE("/:id", h.Review.Delete)

	adminReviews := reviews.Group("")
	adminReviews.Use(middleware.AdminMiddleware())
	adminReviews.GET("", h.Review.GetAll)

	reviewCreate := api.Group("/movies/:id/reviews")
	reviewCreate.Use(middleware.AuthMiddleware(jwtMgr))
	reviewCreate.POST("", h.Review.Create)

	return r
}
