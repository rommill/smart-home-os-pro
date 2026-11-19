package api

import (
	"log"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
)

// [EN] ServiceAuthMiddleware authorizes system services and emulators via X-API-Key without default fallbacks.
// [RU] ServiceAuthMiddleware авторизует системные сервисы и эмуляторы по X-API-Key без дефолтных секретов.
func ServiceAuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		expectedKey := os.Getenv("SERVICE_API_KEY")
		if expectedKey == "" {
			log.Println("❌ [SECURITY ERROR] SERVICE_API_KEY is not configured")
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Service authentication is not configured"})
			c.Abort()
			return
		}

		apiKey := c.GetHeader("X-API-Key")
		if !isValidAPIKey(apiKey, expectedKey) {
			log.Printf("⚠️ [SECURITY WARNING] Unauthorized service request from IP %s", c.ClientIP())
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized service key"})
			c.Abort()
			return
		}

		c.Next()
	}
}