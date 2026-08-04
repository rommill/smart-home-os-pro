package api

import (
	"database/sql"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

// Хэширование пароля перед сохранением в БД
func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(bytes), err
}

// Безопасная проверка пароля
func CheckPasswordHash(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

// Генерация JWT с использованием ENV секрета
func GenerateJWT(userID int) (string, error) {
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		secret = "my_ultra_secret_key_2026"
	}

	claims := jwt.MapClaims{
		"user_id": userID,
		"exp":     time.Now().Add(time.Hour * 24).Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(secret))
}

type LoginInput struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// Login принимает *sql.DB и возвращает gin.HandlerFunc
func Login(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var input LoginInput
		if err := c.ShouldBindJSON(&input); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
			return
		}

		// Заглушка аутентификации с использованием передаваемой БД
		// Позже здесь добавим выборку пользователя из db и проверку через CheckPasswordHash
		_ = db // временно заглушаем неиспользуемую переменную

		c.JSON(http.StatusOK, gin.H{"message": "Login successful"})
	}
}