package api

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"smart-home/models"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// helper function to resolve user identification from request context
func getUserIDFromContext(c *gin.Context) (interface{}, bool) {
	if id, exists := c.Get("user_id"); exists {
		return id, true
	}
	if id, exists := c.Get("userID"); exists {
		return id, true
	}
	if un, exists := c.Get("username"); exists {
		return un, true
	}
	return nil, false
}

// GetTelemetry handles user-specific room data and couples it with MongoDB telemetry streams.
func GetTelemetry(db *sql.DB, mongoColl *mongo.Collection) gin.HandlerFunc {
	return func(c *gin.Context) {
		if db == nil {
			log.Println("❌ [TELEMETRY] Database connection is nil")
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
			return
		}

		userID, exists := getUserIDFromContext(c)
		if !exists || userID == nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized: user context missing"})
			return
		}

		ctx, cancel := context.WithTimeout(c.Request.Context(), 3*time.Second)
		defer cancel()

		tx, err := db.BeginTx(ctx, nil)
		if err != nil {
			log.Printf("❌ [TELEMETRY TX ERROR] BeginTx failed: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to initialize transaction"})
			return
		}
		defer tx.Rollback()

		userIDStr := fmt.Sprintf("%v", userID)
		if _, err := tx.ExecContext(ctx, "SELECT set_config('app.current_user_id', $1, true)", userIDStr); err != nil {
			log.Printf("❌ [TELEMETRY RLS ERROR] Failed to set app.current_user_id: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to configure session context"})
			return
		}

		rows, err := tx.QueryContext(ctx, "SELECT id, name, target_temperature FROM rooms ORDER BY id ASC")
		if err != nil {
			log.Printf("❌ [TELEMETRY DB ERROR] Query failed: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch room records"})
			return
		}
		defer rows.Close()

		results := make([]models.RoomData, 0)

		for rows.Next() {
			var id int
			var rawName string
			var targetTemp sql.NullFloat64

			if err := rows.Scan(&id, &rawName, &targetTemp); err != nil {
				log.Printf("⚠️ [TELEMETRY SCAN ERROR] Row scan failed: %v", err)
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

		if err := rows.Err(); err != nil {
			log.Printf("❌ [TELEMETRY ROWS ERROR] Iteration error: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error processing query results"})
			return
		}

		if err := tx.Commit(); err != nil {
			log.Printf("❌ [TELEMETRY COMMIT ERROR] Transaction commit failed: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to finalize transaction"})
			return
		}

		c.JSON(http.StatusOK, results)
	}
}

// UpdateTargetTemperature safely updates room target temperature parameters with RLS validation.
func UpdateTargetTemperature(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		if db == nil {
			log.Println("❌ [ROOM UPDATE] Database connection is nil")
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
			return
		}

		userID, exists := getUserIDFromContext(c)
		if !exists || userID == nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized: user context missing"})
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
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request payload parameters"})
			return
		}

		ctx, cancel := context.WithTimeout(c.Request.Context(), 3*time.Second)
		defer cancel()

		tx, err := db.BeginTx(ctx, nil)
		if err != nil {
			log.Printf("❌ [ROOM UPDATE TX ERROR] BeginTx failed: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to initialize transaction"})
			return
		}
		defer tx.Rollback()

		userIDStr := fmt.Sprintf("%v", userID)
		if _, err := tx.ExecContext(ctx, "SELECT set_config('app.current_user_id', $1, true)", userIDStr); err != nil {
			log.Printf("❌ [ROOM UPDATE RLS ERROR] Failed to set app.current_user_id: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to configure session context"})
			return
		}

		query := `UPDATE rooms SET target_temperature = $1 WHERE id = $2`
		result, err := tx.ExecContext(ctx, query, input.TargetTemperature, roomID)
		if err != nil {
			log.Printf("❌ [ROOM UPDATE DB ERROR] Query failed: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update room parameters"})
			return
		}

		rowsAffected, err := result.RowsAffected()
		if err != nil || rowsAffected == 0 {
			c.JSON(http.StatusNotFound, gin.H{"error": "Room not found or modification unauthorized"})
			return
		}

		if err := tx.Commit(); err != nil {
			log.Printf("❌ [ROOM UPDATE COMMIT ERROR] Transaction commit failed: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to finalize transaction"})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"message":            "Target temperature updated successfully",
			"room_id":            roomID,
			"target_temperature": input.TargetTemperature,
		})
	}
}

// ReceiveTelemetry receives and stores telemetry payload from BLE bridge into MongoDB.
func ReceiveTelemetry(mongoColl *mongo.Collection) gin.HandlerFunc {
	return func(c *gin.Context) {
		var input models.BLETelemetryInput

		if err := c.ShouldBindJSON(&input); err != nil {
			log.Printf("❌ [BLE INGEST ERROR] Invalid JSON payload: %v", err)
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid payload format"})
			return
		}

		log.Printf("🌡️ [BLE INGEST] Device=%s | Temp=%.2f°C | Hum=%.2f%% | Batt=%d%%",
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

			if _, err := mongoColl.InsertOne(ctx, doc); err != nil {
				log.Printf("❌ [MONGO INGEST ERROR] Failed to insert document: %v", err)
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to record telemetry data"})
				return
			}
		}

		c.JSON(http.StatusOK, gin.H{"status": "success"})
	}
}