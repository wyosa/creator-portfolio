package handlers

import (
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

// NewRateLimiter returns middleware allowing at most max requests per window
// per client IP (fixed window, in-memory). Requests over the limit get a 429
// JSON response.
func NewRateLimiter(max int, window time.Duration) gin.HandlerFunc {
	type counter struct {
		count   int
		resetAt time.Time
	}
	var mu sync.Mutex
	hits := make(map[string]counter)

	return func(c *gin.Context) {
		ip := c.ClientIP()
		now := time.Now()

		mu.Lock()
		// Bound the map: sweep expired entries once it grows large.
		if len(hits) >= 4096 {
			for k, v := range hits {
				if now.After(v.resetAt) {
					delete(hits, k)
				}
			}
		}
		cnt, ok := hits[ip]
		if !ok || now.After(cnt.resetAt) {
			cnt = counter{resetAt: now.Add(window)}
		}
		cnt.count++
		hits[ip] = cnt
		allowed := cnt.count <= max
		mu.Unlock()

		if !allowed {
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{"error": "too many attempts"})
			return
		}
		c.Next()
	}
}
