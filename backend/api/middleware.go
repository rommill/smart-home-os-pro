package api

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	jwt "github.com/golang-jwt/jwt/v5"
)

var (
	jwtSecretOnce sync.Once
	cachedJWTKey  []byte
)

// getJWTSecret retrieves and caches the secret key from environment variables.
// It fails fast (log.Fatalf) if JWT_SECRET is not set.
func getJWTSecret() []byte {
	jwtSecretOnce.Do(func() {
		secret := os.Getenv("JWT_SECRET")
		if secret == "" {
			log.Fatalf("FATAL: JWT_SECRET environment variable is missing")
		}
		cachedJWTKey = []byte(secret)
	})
	return cachedJWTKey
}

// GenerateJWT creates a signed JWT for an authenticated user.
func GenerateJWT(userID interface{}, username, role string) (string, error) {
	jwtKey := getJWTSecret()

	claims := jwt.MapClaims{
		"user_id":  fmt.Sprintf("%v", userID),
		"username": username,
		"role":     role,
		"exp":      time.Now().Add(24 * time.Hour).Unix(),
		"iat":      time.Now().Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(jwtKey)
}

// CORSMiddleware configures cross-origin resource sharing headers securely.
func CORSMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		origin := c.Request.Header.Get("Origin")
		allowedOriginsEnv := os.Getenv("ALLOWED_ORIGIN")
		
		allowedOrigin := "http://localhost:3000"
		if allowedOriginsEnv != "" {
			origins := strings.Split(allowedOriginsEnv, ",")
			for _, o := range origins {
				if strings.TrimSpace(o) == origin {
					allowedOrigin = origin
					break
				}
			}
		} else if origin != "" {
			allowedOrigin = origin
		}

		c.Writer.Header().Set("Access-Control-Allow-Origin", allowedOrigin)
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	}
}

// AuthMiddleware validates the JWT token provided in the Authorization header.
func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization header is required"})
			c.Abort()
			return
		}

		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization format must be Bearer {token}"})
			c.Abort()
			return
		}

		tokenString := parts[1]
		jwtKey := getJWTSecret()

		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}
			return jwtKey, nil
		})

		if err != nil || !token.Valid {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid or expired token"})
			c.Abort()
			return
		}

		if claims, ok := token.Claims.(jwt.MapClaims); ok {
			var userIDStr string
			if val, exists := claims["user_id"]; exists {
				userIDStr = fmt.Sprintf("%v", val)
			}

			// Store both keys for backward compatibility across handlers
			c.Set("userID", userIDStr)
			c.Set("user_id", userIDStr)

			if val, exists := claims["username"]; exists {
				c.Set("username", fmt.Sprintf("%v", val))
			} else {
				c.Set("username", userIDStr)
			}

			if val, exists := claims["role"]; exists {
				c.Set("role", fmt.Sprintf("%v", val))
			}
		}

		c.Next()
	}
}

// RateLimiter struct to track request attempts per IP.
type clientLimiter struct {
	lastSeen time.Time
	count    int
}

var (
	clients     = make(map[string]*clientLimiter)
	clientsMu   sync.Mutex
	cleanupOnce sync.Once
)

// startCleanupRoutine periodically clears stale rate limiter entries to prevent memory leaks.
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

// LoginRateLimiter limits repeated auth requests to prevent brute-force attacks.
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