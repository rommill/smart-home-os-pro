package api

import (
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
)

// [EN] parseAllowedOrigins parses a comma-separated list of exact allowed origins.
// [RU] parseAllowedOrigins разбирает разделённый запятыми список точных разрешённых Origin.
func parseAllowedOrigins(raw string) map[string]struct{} {
	allowed := make(map[string]struct{})
	for _, item := range strings.Split(raw, ",") {
		trimmed := strings.TrimSpace(item)
		if trimmed != "" {
			allowed[trimmed] = struct{}{}
		}
	}
	return allowed
}

// [EN] CORSMiddleware enforces a strict origin allowlist from ALLOWED_ORIGIN without hardcoded fallbacks.
// [RU] CORSMiddleware применяет строгий allowlist из ALLOWED_ORIGIN без зашитых localhost-исключений.
func CORSMiddleware() gin.HandlerFunc {
	rawOrigins := strings.TrimSpace(os.Getenv("ALLOWED_ORIGIN"))
	if rawOrigins == "" {
		log.Println("⚠️ [CORS WARNING] ALLOWED_ORIGIN is empty; cross-origin requests will not be matched")
	}
	allowedOrigins := parseAllowedOrigins(rawOrigins)

	return func(c *gin.Context) {
		origin := strings.TrimSpace(c.Request.Header.Get("Origin"))
		c.Writer.Header().Add("Vary", "Origin")

		if origin != "" {
			if _, ok := allowedOrigins[origin]; ok {
				c.Writer.Header().Set("Access-Control-Allow-Origin", origin)
				c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
				c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
				c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE")
			} else if c.Request.Method == http.MethodOptions {
				c.AbortWithStatus(http.StatusForbidden)
				return
			}
		}

		if c.Request.Method == http.MethodOptions {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	}
}