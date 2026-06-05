package api

import (
	"database/sql"
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
			c.JSON(http.StatusBadRequest, gin.H{"error": "Неверный формат"})
			return
		}

		var userID int
		var dbPassword string
		err := db.QueryRow("SELECT id, password_hash FROM users WHERE username = $1", input.Username).Scan(&userID, &dbPassword)

		if err != nil || dbPassword != input.Password {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Логин или пароль не подошли"})
			return
		}

		token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
			"user_id": userID,
			"exp":     time.Now().Add(time.Hour * 24).Unix(),
		})

		tokenString, _ := token.SignedString(jwtKey)
		c.JSON(http.StatusOK, gin.H{"token": tokenString})
	}
}