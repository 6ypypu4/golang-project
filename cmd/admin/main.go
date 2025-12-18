package main

import (
	"log"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
)

func main() {
	port := os.Getenv("ADMIN_PORT")
	if port == "" {
		port = "8081"
	}

	r := gin.New()
	r.Use(gin.Logger(), gin.Recovery())

	r.GET("/admin/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	if err := r.Run(":" + port); err != nil {
		log.Fatalf("admin server error: %v", err)
	}
}
