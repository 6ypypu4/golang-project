package middleware

import (
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

type rateState struct {
	count     int
	windowEnd time.Time
}

var rateStore = struct {
	mu   sync.Mutex
	data map[string]rateState
}{
	data: make(map[string]rateState),
}

func RateLimit(maxPerMinute int) gin.HandlerFunc {
	window := time.Minute

	return func(c *gin.Context) {
		key := c.ClientIP()
		now := time.Now()

		rateStore.mu.Lock()
		state := rateStore.data[key]
		if state.windowEnd.IsZero() || now.After(state.windowEnd) {
			state.windowEnd = now.Add(window)
			state.count = 0
		}
		state.count++
		rateStore.data[key] = state
		currentCount := state.count
		rateStore.mu.Unlock()

		if currentCount > maxPerMinute {
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{"error": "rate limit exceeded"})
			return
		}

		c.Next()
	}
}


