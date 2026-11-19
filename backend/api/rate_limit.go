package api

import (
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

type clientLimiter struct {
	lastSeen time.Time
	count    int
}

var (
	clients     = make(map[string]*clientLimiter)
	clientsMu   sync.Mutex
	cleanupOnce sync.Once
)

func startCleanupRoutine(window time.Duration) {
	go func() {
		for {
			time.Sleep(window)
			clientsMu.Lock()
			for ip, limiter := range clients {
				if time.Since(limiter.lastSeen) > window {
					delete(clients, ip)
				}
			}
			clientsMu.Unlock()
		}
	}()
}

// [EN] LoginRateLimiter limits repeated auth requests to prevent brute-force attacks.
// [RU] LoginRateLimiter ограничивает частые запросы авторизации для защиты от перебора.
func LoginRateLimiter(maxRequests int, window time.Duration) gin.HandlerFunc {
	cleanupOnce.Do(func() {
		startCleanupRoutine(window)
	})

	return func(c *gin.Context) {
		clientIP := c.ClientIP()

		clientsMu.Lock()
		limiter, exists := clients[clientIP]
		if !exists || time.Since(limiter.lastSeen) > window {
			clients[clientIP] = &clientLimiter{lastSeen: time.Now(), count: 1}
			clientsMu.Unlock()
			c.Next()
			return
		}

		if limiter.count >= maxRequests {
			clientsMu.Unlock()
			c.JSON(http.StatusTooManyRequests, gin.H{"error": "Too many login attempts. Please try again later."})
			c.Abort()
			return
		}

		limiter.count++
		limiter.lastSeen = time.Now()
		clientsMu.Unlock()

		c.Next()
	}
}