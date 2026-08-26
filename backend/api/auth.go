package api

import (
	"database/sql"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
)

type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

func Login(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req LoginRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			log.Printf("❌ [LOGIN ERROR] Invalid JSON request: %v", err)
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid payload"})
			return
		}

		var userID string
		var passwordHash string
		var role string

		// Запрашиваем ID, password_hash и роль (если колонки role нет, по умолчанию "user")
		query := `SELECT id::text, password_hash, COALESCE(role, 'user') FROM users WHERE username = $1`
		err := db.QueryRow(query, req.Username).Scan(&userID, &passwordHash, &role)
		if err != nil {
			if err == sql.ErrNoRows {
				log.Printf("⚠️ [LOGIN] User not found: %s", req.Username)
				c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid username or password"})
				return
			}
			// Если нет колонки role в таблице, сделаем фоллбэк запрос
			queryFallback := `SELECT id::text, password_hash FROM users WHERE username = $1`
			if errFallback := db.QueryRow(queryFallback, req.Username).Scan(&userID, &passwordHash); errFallback != nil {
				log.Printf("❌ [LOGIN DB ERROR] Query failed: %v", errFallback)
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
				return
			}
			role = "user"
		}

		// Проверка пароля
		if err := bcrypt.CompareHashAndPassword([]byte(passwordHash), []byte(req.Password)); err != nil {
			log.Printf("⚠️ [LOGIN] Invalid password for user: %s", req.Username)
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid username or password"})
			return
		}

		// Передаем 3 аргумента в GenerateJWT: (userID, username, role)
		token, err := GenerateJWT(userID, req.Username, role)
		if err != nil {
			log.Printf("❌ [LOGIN JWT ERROR] Failed to generate token: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Token generation error"})
			return
		}

		log.Printf("✅ [LOGIN SUCCESS] User %s logged in successfully (ID: %s)", req.Username, userID)
		c.JSON(http.StatusOK, gin.H{
			"token": token,
			"user": gin.H{
				"id":       userID,
				"username": req.Username,
				"role":     role,
			},
		})
	}
}