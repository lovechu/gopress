package middleware

import (
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/yourorg/gopress/internal/config"
)

type rateLimiter struct {
	requests map[string][]time.Time
	mu       sync.RWMutex
	cfg      config.RateLimitConfig
}

var limiter *rateLimiter

func init() {
	limiter = &rateLimiter{
		requests: make(map[string][]time.Time),
		cfg: config.RateLimitConfig{
			Enabled:  true,
			Requests: 100,
			Window:   60,
		},
	}
}

func RateLimit(cfg config.RateLimitConfig) gin.HandlerFunc {
	if !cfg.Enabled {
		return func(c *gin.Context) {
			c.Next()
		}
	}

	limiter.cfg = cfg

	return func(c *gin.Context) {
		key := c.ClientIP()
		now := time.Now()
		window := time.Duration(cfg.Window) * time.Second

		limiter.mu.Lock()
		times := limiter.requests[key]

		// 清理过期的请求记录
		validTimes := make([]time.Time, 0)
		for _, t := range times {
			if now.Sub(t) < window {
				validTimes = append(validTimes, t)
			}
		}

		if len(validTimes) >= cfg.Requests {
			limiter.mu.Unlock()
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
				"error": "rate limit exceeded",
			})
			return
		}

		validTimes = append(validTimes, now)
		limiter.requests[key] = validTimes
		limiter.mu.Unlock()

		c.Next()
	}
}
