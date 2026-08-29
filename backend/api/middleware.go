package api

import (
	"errors"
	"fmt"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

// getJWTSecret retrieves secret key from environment variables.
func getJWTSecret() ([]byte, error) {
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		return nil, errors.New("JWT_SECRET environment variable is missing")
	}
	return []byte(secret), nil
}

// GenerateJWT creates a signed JWT for an authenticated user.
func GenerateJWT(userID interface{}, username, role string) (string, error) {
	jwtKey, err := getJWTSecret()
	if err != nil {
		return "", err
	}

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

// CORSMiddleware configures cross-origin resource sharing headers.
func CORSMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		allowedOrigin := os.Getenv("ALLOWED_ORIGIN")
		if allowedOrigin == "" {
			allowedOrigin = "http://localhost:3000"
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
		jwtKey, err := getJWTSecret()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Server authentication config error"})
			c.Abort()
			return
		}

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
	clients = make(map[string]*clientLimiter)
	mu      sync.Mutex
)

// LoginRateLimiter limits repeated auth requests to prevent brute-force attacks.
func LoginRateLimiter(maxRequests int, window time.Duration) gin.HandlerFunc {
	return func(c *gin.Context) {
		clientIP := c.ClientIP()

		mu.Lock()
		limiter, exists := clients[clientIP]
		if !exists || time.Since(limiter.lastSeen) > window {
			clients[clientIP] = &clientLimiter{lastSeen: time.Now(), count: 1}
			mu.Unlock()
			c.Next()
			return
		}

		if limiter.count >= maxRequests {
			mu.Unlock()
			c.JSON(http.StatusTooManyRequests, gin.H{"error": "Too many login attempts. Please try again later."})
			c.Abort()
			return
		}

		limiter.count++
		limiter.lastSeen = time.Now()
		mu.Unlock()

		c.Next()
	}
}