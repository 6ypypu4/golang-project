package router

import (
	"github.com/gin-gonic/gin"

	"golang-project/internal/middleware"
)

func New() *gin.Engine {
	r := gin.New()
	r.Use(
		middleware.RequestID(),
		middleware.Logger(),
		middleware.RateLimit(60),
		gin.Recovery(),
		middleware.CORS(),
		middleware.BodyLimit(1<<20),
	)
	return r
}
