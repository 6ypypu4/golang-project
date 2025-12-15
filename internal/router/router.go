package router

import "github.com/gin-gonic/gin"

// New returns a bare Gin engine. For real routes use handler.SetupRoutes in cmd/api.
func New() *gin.Engine {
	return gin.New()
}
