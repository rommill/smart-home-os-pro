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

/**
 * GetTelemetry handles user-specific room data and couples it with MongoDB stream logs.
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

		_, err = tx.Exec("SELECT set_config('app.current_user_id', $1, true)", fmt.Sprintf("%v", userID))
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "RLS context error"})
			return
		}

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
			var targetTemp sql.NullFloat64

			if err := rows.Scan(&id, &nameBytes, &targetTemp); err != nil {
				fmt.Println("❌ SCAN ERROR:", err)
				continue
			}

			nameMap := make(map[string]string)
			if err := json.Unmarshal(nameBytes, &nameMap); err != nil {
				nameMap = map[string]string{"en": string(nameBytes), "ru": string(nameBytes)}
			}

			var lastEntry bson.M
			opts := options.FindOne().SetSort(bson.M{"timestamp": -1})
			
			
			err := mongoColl.FindOne(context.TODO(), bson.M{"device_id": fmt.Sprintf("%d", id)}, opts).Decode(&lastEntry)

			temp := "N/A"
			lastTime := time.Time{}
			if err == nil {
				temp = fmt.Sprintf("%v", lastEntry["value"])
				if t, ok := lastEntry["timestamp"].(primitive.DateTime); ok {
					lastTime = t.Time()
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
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Transaction commit error"})
			return
		}

		c.JSON(http.StatusOK, results)
	}
}

/**
 * UpdateTargetTemperature safely updates room temperature parameters
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

		var input struct {
			TargetTemperature float64 `json:"target_temperature"`
		}

		if err := c.ShouldBindJSON(&input); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid payload parameters"})
			return
		}

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

/**
 * ReceiveTelemetry принимает данные от BLE
 */
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
			doc := bson.M{
				"device_id":   input.DeviceID,
				"value":       input.Temperature,
				"humidity":    input.Humidity,
				"battery":     input.Battery,
				"sensor_type": input.SensorType,
				"rssi":        input.RSSI,
				"timestamp":   primitive.NewDateTimeFromTime(time.Unix(input.Timestamp, 0)),
			}
			_, _ = mongoColl.InsertOne(context.TODO(), doc)
		}

		c.JSON(http.StatusOK, gin.H{"status": "success"})
	}
}