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

// [EN] getJWTSecret retrieves and caches the secret key from environment variables.
// [RU] getJWTSecret читает и кэширует секретный ключ из переменных окружения, падая при отсутствии.
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

// [EN] GenerateJWT creates a signed JWT for an authenticated user.
// [RU] GenerateJWT создаёт подписанный JWT для аутентифицированного пользователя.
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

// [EN] AuthMiddleware validates the JWT token provided in the Authorization header.
// [RU] AuthMiddleware проверяет JWT-токен из заголовка Authorization.
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