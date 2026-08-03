package api

import (
	"database/sql"
	"fmt" // Required for internal English server logging
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

func Login(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var input struct {
			Username string `json:"username"`
			Password string `json:"password"`
		}

		if err := c.ShouldBindJSON(&input); err != nil {
			// Context: Invalid JSON structure received from the frontend client
			fmt.Println("❌ API Error: Invalid login payload structure:", err)
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid payload format"})
			return
		}

		var userID int
		var dbPassword string
		
		// Querying users table initialized by the Spring Boot framework
		err := db.QueryRow("SELECT id, password_hash FROM users WHERE username = $1", input.Username).Scan(&userID, &dbPassword)

		if err != nil {
			if err == sql.ErrNoRows {
				// Context: PostgreSQL returned empty result set for this username
				fmt.Printf("❌ Auth Fail: User '%s' not found in PostgreSQL database\n", input.Username)
			} else {
				// Context: Database connection drop or schema mismatch issues
				fmt.Println("❌ Database Error: Failed to execute user query:", err)
			}
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid username or password"})
			return
		}

		// Plain text string comparison used for initial testing phase
		if dbPassword != input.Password {
			fmt.Printf("❌ Auth Fail: Password mismatch for user '%s'\n", input.Username)
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid username or password"})
			return
		}

		// Context: Credentials successfully validated, safe to generate session JWT
		fmt.Printf("✅ Auth Success: User '%s' (ID: %d) authenticated successfully\n", input.Username, userID)

		token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
			"user_id": userID,
			"exp":     time.Now().Add(time.Hour * 24).Unix(),
		})

		tokenString, _ := token.SignedString(jwtKey)
		c.JSON(http.StatusOK, gin.H{"token": tokenString})
	}
}