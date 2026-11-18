package api

import (
	"context"
	"log"
	"net/http"
	"os"
	"time"

	"smart-home/models"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/mongo"
)

// [EN] PostTelemetry handles incoming telemetry payloads from climate emulators and BLE bridges with API-key protection.
// [RU] PostTelemetry обрабатывает входящую телеметрию от климатических эмуляторов и BLE-мостов с защитой по API-ключу.
func PostTelemetry(mongoColl *mongo.Collection) gin.HandlerFunc {
	return func(c *gin.Context) {
		// [EN] Retrieve expected API key from environment variables
		// [RU] Получаем ожидаемый API-ключ из переменных окружения
		expectedAPIKey := os.Getenv("SERVICE_API_KEY")
		if expectedAPIKey == "" {
			expectedAPIKey = "secret_emulator_key_123" // Fallback default
		}

		// [EN] Validate X-API-Key header to prevent unauthorized telemetry injection
		// [RU] Проверяем заголовок X-API-Key для защиты от несанкционированной отправки телеметрии
		clientKey := c.GetHeader("X-API-Key")
		if clientKey != expectedAPIKey {
			log.Printf("⚠️ [SECURITY WARNING] Unauthorized telemetry attempt with key: %s", clientKey)
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized: invalid or missing API key"})
			return
		}

		if mongoColl == nil {
			log.Println("❌ [TELEMETRY ERROR] MongoDB collection is nil")
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Database not initialized"})
			return
		}

		var input models.BLETelemetryInput
		if err := c.ShouldBindJSON(&input); err != nil {
			log.Printf("⚠️ [TELEMETRY PARSE ERROR] Invalid request body: %v", err)
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request payload"})
			return
		}

		ctx, cancel := context.WithTimeout(c.Request.Context(), 3*time.Second)
		defer cancel()

		// [EN] Insert raw telemetry document into MongoDB
		// [RU] Сохраняем сырой документ телеметрии в MongoDB
		doc := map[string]interface{}{
			"device_id":   input.DeviceID,
			"rssi":        input.RSSI,
			"timestamp":   input.Timestamp,
			"sensor_type": input.SensorType,
			"value":       input.Temperature, // Совместимость с текущими графиками
			"temperature": input.Temperature,
			"humidity":    input.Humidity,
			"battery":     input.Battery,
			"voltage_mv":  input.VoltageMV,
			"created_at":  time.Now(),
		}

		_, err := mongoColl.InsertOne(ctx, doc)
		if err != nil {
			log.Printf("❌ [TELEMETRY DB ERROR] Failed to insert document: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to store telemetry data"})
			return
		}

		c.JSON(http.StatusCreated, gin.H{
			"status":  "success",
			"message": "Telemetry recorded successfully",
		})
	}
}