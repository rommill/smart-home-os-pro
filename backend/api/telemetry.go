package api

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"smart-home/models"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// GetTelemetry handles user-specific room data and couples it with MongoDB stream logs.
func GetTelemetry(db *sql.DB, mongoColl *mongo.Collection) gin.HandlerFunc {
	return func(c *gin.Context) {
		if db == nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Database connection is nil"})
			return
		}

		
		var userID interface{}
		if id, exists := c.Get("user_id"); exists {
			userID = id
		} else if id, exists := c.Get("userID"); exists {
			userID = id
		} else if un, exists := c.Get("username"); exists {
			userID = un
		}

		if userID == nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized: user context missing"})
			return
		}

		ctx, cancel := context.WithTimeout(c.Request.Context(), 3*time.Second)
		defer cancel()

		tx, err := db.BeginTx(ctx, nil)
		if err != nil {
			fmt.Println("❌ BEGIN TX ERROR:", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Transaction start error"})
			return
		}
		defer tx.Rollback()

	
		userIDStr := fmt.Sprintf("%v", userID)
		_, err = tx.ExecContext(ctx, "SELECT set_config('app.current_user_id', $1, true)", userIDStr)
		if err != nil {
			fmt.Println("❌ RLS SET_CONFIG ERROR:", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "RLS context error"})
			return
		}

		rows, err := tx.QueryContext(ctx, "SELECT id, name, target_temperature FROM rooms ORDER BY id ASC")
		if err != nil {
			fmt.Println("❌ POSTGRESQL QUERY ERROR:", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "DB query error"})
			return
		}
		defer rows.Close()

		results := make([]models.RoomData, 0)

		for rows.Next() {
			var id int
			var rawName string
			var targetTemp sql.NullFloat64

			if err := rows.Scan(&id, &rawName, &targetTemp); err != nil {
				fmt.Println("❌ SCAN ERROR:", err)
				continue
			}

			
			nameMap := make(map[string]string)
			if err := json.Unmarshal([]byte(rawName), &nameMap); err != nil {
				nameMap = map[string]string{"en": rawName, "ru": rawName}
			}

			temp := "N/A"
			lastTime := time.Time{}

			
			if mongoColl != nil {
				var lastEntry bson.M
				opts := options.FindOne().SetSort(bson.M{"timestamp": -1})

				err := mongoColl.FindOne(ctx, bson.M{"device_id": fmt.Sprintf("%d", id)}, opts).Decode(&lastEntry)
				if err == nil {
					if val, ok := lastEntry["value"]; ok {
						temp = fmt.Sprintf("%v", val)
					}
					if t, ok := lastEntry["timestamp"].(primitive.DateTime); ok {
						lastTime = t.Time()
					}
				}
			}

			finalTargetTemp := 23.0
			if targetTemp.Valid {
				finalTargetTemp = targetTemp.Float64
			}

			results = append(results, models.RoomData{
				ID:                id,
				Name:              nameMap,
				Temperature:       temp,
				LastUpdate:        lastTime,
				TargetTemperature: finalTargetTemp,
			})
		}

		if err := tx.Commit(); err != nil {
			fmt.Println("❌ COMMIT ERROR:", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Transaction commit error"})
			return
		}

		c.JSON(http.StatusOK, results)
	}
}

// UpdateTargetTemperature safely updates room temperature parameters
func UpdateTargetTemperature(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		if db == nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Database connection is nil"})
			return
		}

		var userID interface{}
		if id, exists := c.Get("user_id"); exists {
			userID = id
		} else if id, exists := c.Get("userID"); exists {
			userID = id
		} else if un, exists := c.Get("username"); exists {
			userID = un
		}

		if userID == nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized context missing"})
			return
		}

		roomIDStr := c.Param("id")
		var roomID int
		if _, err := fmt.Sscanf(roomIDStr, "%d", &roomID); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid room ID format"})
			return
		}

		var input struct {
			TargetTemperature float64 `json:"target_temperature"`
		}

		if err := c.ShouldBindJSON(&input); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid payload parameters"})
			return
		}

		ctx, cancel := context.WithTimeout(c.Request.Context(), 3*time.Second)
		defer cancel()

		tx, err := db.BeginTx(ctx, nil)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Transaction start error"})
			return
		}
		defer tx.Rollback()

		userIDStr := fmt.Sprintf("%v", userID)
		_, err = tx.ExecContext(ctx, "SELECT set_config('app.current_user_id', $1, true)", userIDStr)
		if err != nil {
			fmt.Println("❌ RLS SET_CONFIG ERROR:", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "RLS context error"})
			return
		}

		query := `UPDATE rooms SET target_temperature = $1 WHERE id = $2`
		result, err := tx.ExecContext(ctx, query, input.TargetTemperature, roomID)
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

// ReceiveTelemetry receives telemetry payload from BLE bridge
func ReceiveTelemetry(mongoColl *mongo.Collection) gin.HandlerFunc {
	return func(c *gin.Context) {
		var input models.BLETelemetryInput

		if err := c.ShouldBindJSON(&input); err != nil {
			fmt.Println("❌ INVALID JSON FROM BLE BRIDGE:", err)
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid payload format"})
			return
		}

		fmt.Printf("🌡️ [GO API] Received BLE Telemetry: Device=%s | Temp=%.2f°C | Hum=%.2f%% | Batt=%d%%\n",
			input.DeviceID, input.Temperature, input.Humidity, input.Battery)

		if mongoColl != nil {
			ctx, cancel := context.WithTimeout(c.Request.Context(), 3*time.Second)
			defer cancel()

			doc := bson.M{
				"device_id":   input.DeviceID,
				"value":       input.Temperature,
				"humidity":    input.Humidity,
				"battery":     input.Battery,
				"sensor_type": input.SensorType,
				"rssi":        input.RSSI,
				"timestamp":   primitive.NewDateTimeFromTime(time.Unix(input.Timestamp, 0)),
			}
			_, _ = mongoColl.InsertOne(ctx, doc)
		}

		c.JSON(http.StatusOK, gin.H{"status": "success"})
	}
}