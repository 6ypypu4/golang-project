package middleware

import (
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
)

// Logger produces structured log lines with request metadata.
func Logger() gin.HandlerFunc {
	return gin.LoggerWithFormatter(func(param gin.LogFormatterParams) string {
		reqID := param.Request.Header.Get(requestIDHeader)
		if reqID == "" {
			if val, ok := param.Keys[string(ContextUserID)]; ok {
				if s, ok := val.(string); ok {
					reqID = s
				}
			}
		}
		return fmt.Sprintf(`{"time":"%s","method":"%s","path":"%s","status":%d,"latency":"%s","ip":"%s","user_agent":"%s","request_id":"%s"}`+"\n",
			param.TimeStamp.Format(time.RFC3339),
			param.Method,
			param.Path,
			param.StatusCode,
			param.Latency,
			param.ClientIP,
			param.Request.UserAgent(),
			reqID,
		)
	})
}
