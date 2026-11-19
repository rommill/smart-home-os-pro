package api

import (
	"context"
	"crypto/subtle"
	"errors"
	"log"
	"math"
	"net/http"
	"os"
	"strings"
	"time"

	"smart-home/models"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/mongo"
)

// [EN] isValidAPIKey performs a constant-time comparison between client and expected API keys.
// [RU] isValidAPIKey выполняет сравнение клиентского и ожидаемого API-ключей за постоянное время.
func isValidAPIKey(clientKey, expectedKey string) bool {
	if expectedKey == "" || clientKey == "" {
		return false
	}
	return subtle.ConstantTimeCompare([]byte(clientKey), []byte(expectedKey)) == 1
}

// [EN] validateTelemetryInput validates core fields and physical bounds of incoming telemetry.
// [RU] validateTelemetryInput проверяет обязательные поля и физические границы входящей телеметрии.
func validateTelemetryInput(input models.BLETelemetryInput) error {
	deviceID := strings.TrimSpace(input.DeviceID)
	if deviceID == "" {
		return errors.New("device_id is required")
	}

	if math.IsNaN(input.Temperature) || math.IsInf(input.Temperature, 0) {
		return errors.New("temperature must be a finite number")
	}
	if input.Temperature < -50 || input.Temperature > 120 {
		return errors.New("temperature is out of valid range (-50 to 120)")
	}

	if math.IsNaN(input.Humidity) || math.IsInf(input.Humidity, 0) {
		return errors.New("humidity must be a finite number")
	}
	if input.Humidity < 0 || input.Humidity > 100 {
		return errors.New("humidity is out of valid range (0 to 100)")
	}

	if input.Battery < 0 || input.Battery > 100 {
		return errors.New("battery is out of valid range (0 to 100)")
	}

	return nil
}

// [EN] PostTelemetry handles incoming telemetry payloads from climate emulators and BLE bridges with API-key protection.
// [RU] PostTelemetry обрабатывает входящую телеметрию от климатических эмуляторов и BLE-мостов с защитой по API-ключу.
func PostTelemetry(mongoColl *mongo.Collection) gin.HandlerFunc {
	return func(c *gin.Context) {
		// [EN] Read expected API key from environment without insecure fallback defaults
		// [RU] Читаем ожидаемый API-ключ из окружения без небезопасных дефолтных значений
		expectedAPIKey := os.Getenv("SERVICE_API_KEY")
		if expectedAPIKey == "" {
			log.Println("❌ [SECURITY ERROR] SERVICE_API_KEY is not configured")
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Service authentication is not configured"})
			return
		}

		// [EN] Validate X-API-Key header using constant-time comparison without logging secrets
		// [RU] Проверяем заголовок X-API-Key через constant-time сравнение без логирования секретов
		clientKey := c.GetHeader("X-API-Key")
		if !isValidAPIKey(clientKey, expectedAPIKey) {
			log.Printf("⚠️ [SECURITY WARNING] Unauthorized telemetry attempt from IP %s", c.ClientIP())
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

		// [EN] Reject out-of-range or malformed sensor values before writing to MongoDB
		// [RU] Отклоняем некорректные или выходящие за границы значения датчиков до записи в MongoDB
		if err := validateTelemetryInput(input); err != nil {
			log.Printf("⚠️ [TELEMETRY VALIDATION ERROR] device=%s: %v", strings.TrimSpace(input.DeviceID), err)
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		ctx, cancel := context.WithTimeout(c.Request.Context(), 3*time.Second)
		defer cancel()

		// [EN] Insert raw telemetry document into MongoDB
		// [RU] Сохраняем сырой документ телеметрии в MongoDB
		doc := map[string]interface{}{
			"device_id":   strings.TrimSpace(input.DeviceID),
			"rssi":        input.RSSI,
			"timestamp":   input.Timestamp,
			"sensor_type": input.SensorType,
			"value":       input.Temperature,
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