package middleware

import (
	"github.com/gin-gonic/gin"
	"math/rand"
	"strconv"
	"time"
)

const requestIDHeader = "X-Request-ID"

func RequestID() gin.HandlerFunc {
	rand.Seed(time.Now().UnixNano())
	return func(c *gin.Context) {
		id := c.GetHeader(requestIDHeader)
		if id == "" {
			id = strconv.Itoa(rand.Intn(1_000_000_000)) // случайный int в виде строки
		}
		c.Writer.Header().Set(requestIDHeader, id)
		c.Next()
	}
}
