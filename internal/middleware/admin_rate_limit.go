package middleware

import (
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"

	"todo/internal/utils"
)

// AdminLoginRateLimit limits token attempts to maxAttempts within the given window.
// Uses a single global counter since the admin interface is localhost-only.
func AdminLoginRateLimit(maxAttempts int, window time.Duration) gin.HandlerFunc {
	var mu sync.Mutex
	attempts := make([]time.Time, 0, maxAttempts)

	return func(c *gin.Context) {
		now := time.Now()
		cutoff := now.Add(-window)

		mu.Lock()
		valid := attempts[:0]
		for _, t := range attempts {
			if t.After(cutoff) {
				valid = append(valid, t)
			}
		}
		attempts = valid

		if len(attempts) >= maxAttempts {
			mu.Unlock()
			utils.RespondLocalizedError(c, http.StatusTooManyRequests, "rate_limited_admin_login")
			c.Abort()
			return
		}
		attempts = append(attempts, now)
		mu.Unlock()
		c.Next()
	}
}
