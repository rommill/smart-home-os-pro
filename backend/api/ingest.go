package api

import (
    "context"
    "log"
    "net/http"
    "time"

    "smart-home/metrics"
    "smart-home/models"

    "github.com/gin-gonic/gin"
    "go.mongodb.org/mongo-driver/bson"
    "go.mongodb.org/mongo-driver/bson/primitive"
    "go.mongodb.org/mongo-driver/mongo"
)

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

        metrics.ProcessedTelemetryTotal.Inc()

        c.JSON(http.StatusOK, gin.H{"status": "success"})
    }
}