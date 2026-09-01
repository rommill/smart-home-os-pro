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
			log.Printf("❌ [LOGIN ERROR] Invalid JSON payload: %v", err)
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request payload"})
			return
		}

		var userID string
		var passwordHash string
		var role string

		// Query user details: ID, password_hash, and role (defaults to 'user' if role column is missing or null)
		query := `SELECT id::text, password_hash, COALESCE(role, 'user') FROM users WHERE username = $1`
		err := db.QueryRowContext(c.Request.Context(), query, req.Username).Scan(&userID, &passwordHash, &role)
		if err != nil {
			if err == sql.ErrNoRows {
				log.Printf("⚠️ [LOGIN] User not found: %s", req.Username)
				c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid username or password"})
				return
			}

			// Fallback query in case the 'role' column does not exist in the table schema
			queryFallback := `SELECT id::text, password_hash FROM users WHERE username = $1`
			if errFallback := db.QueryRowContext(c.Request.Context(), queryFallback, req.Username).Scan(&userID, &passwordHash); errFallback != nil {
				if errFallback == sql.ErrNoRows {
					log.Printf("⚠️ [LOGIN] User not found (fallback query): %s", req.Username)
					c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid username or password"})
					return
				}
				log.Printf("❌ [LOGIN DB ERROR] Query failed: %v", errFallback)
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
				return
			}
			role = "user"
		}

		// Securely compare stored bcrypt hash with the provided plain text password
		if err := bcrypt.CompareHashAndPassword([]byte(passwordHash), []byte(req.Password)); err != nil {
			log.Printf("⚠️ [LOGIN] Invalid password for user: %s", req.Username)
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid username or password"})
			return
		}

		// Generate JWT with claims: (userID, username, role)
		token, err := GenerateJWT(userID, req.Username, role)
		if err != nil {
			log.Printf("❌ [LOGIN JWT ERROR] Failed to generate token: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to authenticate user"})
			return
		}

		log.Printf("✅ [LOGIN SUCCESS] User %s logged in successfully (ID: %s, Role: %s)", req.Username, userID, role)
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