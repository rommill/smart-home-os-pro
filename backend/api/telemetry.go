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

// [EN] GetTelemetry handles user-specific room data and couples it with active MongoDB telemetry streams.
// [RU] GetTelemetry обрабатывает данные комнат пользователя и объединяет их с активными потоками телеметрии из MongoDB.
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
        defer func() {
            _ = tx.Rollback()
        }()

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

        roomsResults := make([]models.RoomData, 0)

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

            roomsResults = append(roomsResults, models.RoomData{
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

        // --- СЕКЦИЯ ПОИСКА АКТИВНЫХ ФИЗИЧЕСКИХ ДАТЧИКОВ (XIAOMI / BLE) ИЗ MONGODB ---
        activeDevices := make([]models.DeviceData, 0)
        if mongoColl != nil {
            threeMinutesAgo := time.Now().Add(-3 * time.Minute)
            
            pipeline := mongo.Pipeline{
                {{Key: "$sort", Value: bson.M{"timestamp": -1}}},
                {{Key: "$group", Value: bson.M{
                    "_id":         "$device_id",
                    "temperature": bson.M{"$first": "$value"},
                    "humidity":    bson.M{"$first": "$humidity"},
                    "battery":     bson.M{"$first": "$battery"},
                    "sensor_type": bson.M{"$first": "$sensor_type"},
                    "rssi":        bson.M{"$first": "$rssi"},
                    "timestamp":   bson.M{"$first": "$timestamp"},
                }}},
                {{Key: "$match", Value: bson.M{
                    "timestamp": bson.M{"$gte": primitive.NewDateTimeFromTime(threeMinutesAgo)},
                }}},
            }

            cursor, err := mongoColl.Aggregate(ctx, pipeline)
            if err == nil {
                defer cursor.Close(ctx)
                for cursor.Next(ctx) {
                    var rawDoc struct {
                        ID          string             `bson:"_id"`
                        Temperature float64            `bson:"temperature"`
                        Humidity    float64            `bson:"humidity"`
                        Battery     int                `bson:"battery"`
                        SensorType  string             `bson:"sensor_type"`
                        RSSI        int                `bson:"rssi"`
                        Timestamp   primitive.DateTime `bson:"timestamp"`
                    }

                    if err := cursor.Decode(&rawDoc); err == nil {
                        if rawDoc.ID != "1" && rawDoc.ID != "2" {
                            activeDevices = append(activeDevices, models.DeviceData{
                                DeviceID:    rawDoc.ID,
                                Temperature: rawDoc.Temperature,
                                Humidity:    rawDoc.Humidity,
                                Battery:     rawDoc.Battery,
                                SensorType:  rawDoc.SensorType,
                                RSSI:        rawDoc.RSSI,
                                LastUpdate:  rawDoc.Timestamp.Time(),
                            })
                        }
                    }
                }
            }
        }

        c.JSON(http.StatusOK, gin.H{
            "rooms":   roomsResults,
            "devices": activeDevices,
        })
    }
}