package middleware

import (
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"golang.org/x/time/rate"
)

type limiterEntry struct {
	limiter    *rate.Limiter
	lastAccess time.Time
}

type RateLimiter struct {
	limiters *sync.Map
	rate     rate.Limit
	burst    int
	logger   *zap.Logger
	stopCh   chan struct{}
	stopOnce sync.Once
}

func NewRateLimiter(rateLimit int, logger *zap.Logger) *RateLimiter {
	rps := float64(rateLimit) / 60.0
	if rps < 1 {
		rps = 1
	}

	rl := &RateLimiter{
		limiters: &sync.Map{},
		rate:     rate.Limit(rps),
		burst:    rateLimit,
		logger:   logger,
		stopCh:   make(chan struct{}),
	}

	go rl.cleanupLimiters()

	return rl
}

func (rl *RateLimiter) getLimiter(ip string) *rate.Limiter {
	value, exists := rl.limiters.Load(ip)
	if !exists {
		newLimiter := rate.NewLimiter(rl.rate, rl.burst)
		entry := &limiterEntry{
			limiter:    newLimiter,
			lastAccess: time.Now(),
		}
		rl.limiters.Store(ip, entry)
		return newLimiter
	}
	entry := value.(*limiterEntry)
	entry.lastAccess = time.Now()
	return entry.limiter
}

func (rl *RateLimiter) cleanupLimiters() {
	ticker := time.NewTicker(10 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			now := time.Now()
			cutoffTime := now.Add(-10 * time.Minute)
			var keysToDelete []interface{}

			rl.limiters.Range(func(key, value interface{}) bool {
				entry := value.(*limiterEntry)
				if entry.lastAccess.Before(cutoffTime) {
					keysToDelete = append(keysToDelete, key)
				}
				return true
			})

			for _, key := range keysToDelete {
				rl.limiters.Delete(key)
			}

			if len(keysToDelete) > 0 {
				rl.logger.Debug("Cleaned up unused rate limiters",
					zap.Int("count", len(keysToDelete)),
				)
			}
		case <-rl.stopCh:
			rl.logger.Info("Rate limiter cleanup goroutine stopped")
			return
		}
	}
}

func (rl *RateLimiter) Shutdown() {
	rl.stopOnce.Do(func() {
		close(rl.stopCh)
	})
}

func (rl *RateLimiter) Middleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.Request.URL.Path == "/health" {
			c.Next()
			return
		}

		ip := c.ClientIP()

		limiter := rl.getLimiter(ip)

		if !limiter.Allow() {
			rl.logger.Warn("Rate limit exceeded",
				zap.String("ip", ip),
				zap.String("path", c.Request.URL.Path),
				zap.String("method", c.Request.Method),
			)

			requestID := GetRequestID(c)
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error":      "Rate limit exceeded",
				"message":    "Too many requests. Please try again later.",
				"request_id": requestID,
			})
			c.Abort()
			return
		}

		c.Next()
	}
}
