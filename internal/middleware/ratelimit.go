package middleware

import (
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"

	"todo/internal/utils"
)

type tokenBucket struct {
	tokens     float64
	maxTokens  float64
	refillRate float64 // tokens per second
	lastRefill time.Time
	mu         sync.Mutex
}

func (b *tokenBucket) allow() bool {
	b.mu.Lock()
	defer b.mu.Unlock()

	now := time.Now()
	elapsed := now.Sub(b.lastRefill).Seconds()
	b.tokens += elapsed * b.refillRate
	if b.tokens > b.maxTokens {
		b.tokens = b.maxTokens
	}
	b.lastRefill = now

	if b.tokens < 1 {
		return false
	}
	b.tokens--
	return true
}

// IPRateLimit 基于客户端 IP 的令牌桶限流。
// maxReqPerSec 为每秒补充的令牌数，burst 为桶容量（允许的瞬时峰值）。
func IPRateLimit(maxReqPerSec float64, burst int) gin.HandlerFunc {
	buckets := make(map[string]*tokenBucket)
	var mu sync.Mutex

	return func(c *gin.Context) {
		ip := c.ClientIP()

		mu.Lock()
		b, ok := buckets[ip]
		if !ok {
			b = &tokenBucket{
				tokens:     float64(burst),
				maxTokens:  float64(burst),
				refillRate: maxReqPerSec,
				lastRefill: time.Now(),
			}
			buckets[ip] = b
		}
		mu.Unlock()

		if !b.allow() {
			utils.RespondLocalizedError(c, http.StatusTooManyRequests, "rate_limited")
			c.Abort()
			return
		}
		c.Next()
	}
}
