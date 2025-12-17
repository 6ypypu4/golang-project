package middleware

import (
	"crypto/rand"
	"encoding/binary"
	"github.com/gin-gonic/gin"
	"strconv"
)

const requestIDHeader = "X-Request-ID"

func RequestID() gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.GetHeader(requestIDHeader)
		if id == "" {
			var b [8]byte
			if _, err := rand.Read(b[:]); err != nil {
				c.Writer.Header().Set(requestIDHeader, "0")
			} else {
				randomInt := int(binary.LittleEndian.Uint64(b[:]) & 0x7FFFFFFF)
				c.Writer.Header().Set(requestIDHeader, strconv.Itoa(randomInt))
			}
		} else {
			c.Writer.Header().Set(requestIDHeader, id)
		}
		c.Next()
	}
}
