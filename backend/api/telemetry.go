package api

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"smart-home/models"
	"time"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

/**
 * GetTelemetry handles user-specific room data and couples it with MongoDB stream logs.
 * Enforces deterministic sorting to prevent UI layout shifts during runtime polling cycles.
 */
func GetTelemetry(db *sql.DB, mongoColl *mongo.Collection) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, _ := c.Get("userID")

		tx, err := db.Begin()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Transaction start error"})
			return
		}
		defer tx.Rollback()

		// Secure RLS context binding for PostgreSQL isolation rules
		_, err = tx.Exec("SELECT set_config('app.current_user_id', $1, true)", fmt.Sprintf("%v", userID))
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "RLS context error"})
			return
		}

		// FIXED: Added ORDER BY id ASC to preserve strict card order on the frontend UI
		rows, err := tx.Query("SELECT id, name, target_temperature FROM rooms ORDER BY id ASC")
		if err != nil {
			fmt.Println("❌ POSTGRESQL QUERY ERROR:", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "DB error"})
			return
		}
		defer rows.Close()

		var results []models.RoomData

		for rows.Next() {
			var id int
			var nameBytes []byte
			var targetTemp sql.NullFloat64 // Using NullFloat64 in case database field has null values

			if err := rows.Scan(&id, &nameBytes, &targetTemp); err != nil {
				fmt.Println("❌ SCAN ERROR:", err)
				continue
			}

			nameMap := make(map[string]string)
			if err := json.Unmarshal(nameBytes, &nameMap); err != nil {
				nameMap = map[string]string{"en": string(nameBytes), "ru": string(nameBytes)}
			}

			// High frequency monitoring node extraction out of Mongo cluster logs
			var lastEntry bson.M
			opts := options.FindOne().SetSort(bson.M{"timestamp": -1})
			err := mongoColl.FindOne(context.TODO(), bson.M{"device_id": id}, opts).Decode(&lastEntry)

			temp := "N/A"
			lastTime := time.Time{}
			if err == nil {
				temp = fmt.Sprintf("%v", lastEntry["value"])
				if t, ok := lastEntry["timestamp"].(primitive.DateTime); ok {
					lastTime = t.Time()
				}
			}

			// Fallback to default 23.0 if database returns NULL/empty
			finalTargetTemp := 23.0
			if targetTemp.Valid {
				finalTargetTemp = targetTemp.Float64
			}

			// Populating unified structural domain entities back to client views
			results = append(results, models.RoomData{
				ID:                id,
				Name:              nameMap,
				Temperature:       temp,
				LastUpdate:        lastTime,
				TargetTemperature: finalTargetTemp,
			})
		}

		if err := tx.Commit(); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Transaction commit error"})
			return
		}

		c.JSON(http.StatusOK, results)
	}
}

/**
 * UpdateTargetTemperature safely updates room temperature parameters using implicit RLS policies
 */
func UpdateTargetTemperature(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, _ := c.Get("userID")
		roomIDStr := c.Param("id")

		var roomID int
		if _, err := fmt.Sscanf(roomIDStr, "%d", &roomID); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid room ID format"})
			return
		}

		// FIXED: Removed strict binding required check to handle float numeric allocations cleanly
		var input struct {
			TargetTemperature float64 `json:"target_temperature"`
		}

		if err := c.ShouldBindJSON(&input); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid payload parameters"})
			return
		}

		// Using full RLS boundaries for write transactions to maintain absolute isolation
		tx, err := db.Begin()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Transaction start error"})
			return
		}
		defer tx.Rollback()

		_, err = tx.Exec("SELECT set_config('app.current_user_id', $1, true)", fmt.Sprintf("%v", userID))
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "RLS context error"})
			return
		}

		// Updating parameters securely. RLS policy will automatically reject changes if this room doesn't belong to the user
		query := `UPDATE rooms SET target_temperature = $1 WHERE id = $2`
		result, err := tx.Exec(query, input.TargetTemperature, roomID)
		if err != nil {
			fmt.Println("❌ UPDATE EXECUTION ERROR:", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update database state"})
			return
		}

		rowsAffected, err := result.RowsAffected()
		if err != nil || rowsAffected == 0 {
			c.JSON(http.StatusNotFound, gin.H{"error": "Target room resource not found or unauthorized"})
			return
		}

		if err := tx.Commit(); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Transaction commit error"})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"message":            "Target temperature updated successfully",
			"room_id":            roomID,
			"target_temperature": input.TargetTemperature,
		})
	}
}