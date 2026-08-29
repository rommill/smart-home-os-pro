package mqtt

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	"go.mongodb.org/mongo-driver/mongo"
)

// TelemetryPayload represents incoming sensor metrics from devices
type TelemetryPayload struct {
	DeviceID   string                 `json:"device_id" bson:"device_id"`
	SensorType string                 `json:"sensor_type" bson:"sensor_type"`
	Value      float64                `json:"value" bson:"value"`
	Unit       string                 `json:"unit" bson:"unit"`
	Timestamp  int64                  `json:"timestamp" bson:"timestamp"`
	Metadata   map[string]interface{} `json:"metadata,omitempty" bson:"metadata,omitempty"`
}

var (
	mongoCollection *mongo.Collection
	postgresDB      *sql.DB
)

var messagePubHandler mqtt.MessageHandler = func(client mqtt.Client, msg mqtt.Message) {
	var payload TelemetryPayload
	err := json.Unmarshal(msg.Payload(), &payload)
	if err != nil {
		log.Printf("⚠️ [MQTT Error] Failed to parse payload from topic %s: %v", msg.Topic(), err)
		return
	}

	if payload.Timestamp == 0 {
		payload.Timestamp = time.Now().Unix()
	}

	log.Printf("📡 [MQTT Received] Topic: %s | Device: %s | %s: %.2f %s",
		msg.Topic(), payload.DeviceID, payload.SensorType, payload.Value, payload.Unit)

	// 1. Сохранение сырой телеметрии в MongoDB (история)
	if mongoCollection != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		_, err := mongoCollection.InsertOne(ctx, payload)
		if err != nil {
			log.Printf("❌ [MongoDB Error] Failed to insert telemetry: %v", err)
		} else {
			log.Printf("💾 [MongoDB] Ingested telemetry for device: %s", payload.DeviceID)
		}
	}

	// 2. Обновление текущего состояния комнаты в PostgreSQL (для фронтенда)
	if postgresDB != nil && payload.SensorType == "temperature" {
		// Извлекаем имя комнаты из топика "home/Living Room/telemetry"
		parts := strings.Split(msg.Topic(), "/")
		if len(parts) >= 2 {
			roomName := parts[1] // "Living Room" или "Bedroom"
			tempStr := fmt.Sprintf("%.1f", payload.Value)

			// Ищем комнату по имени (в JSON-поле name->>'en' или name->>'ru' или полю name)
			query := `
				UPDATE rooms 
				SET temperature = $1, last_update = NOW() 
				WHERE name->>'en' = $2 OR name->>'ru' = $2
			`
			res, err := postgresDB.Exec(query, tempStr, roomName)
			if err != nil {
				log.Printf("❌ [Postgres Error] Failed to update room %s: %v", roomName, err)
			} else {
				rows, _ := res.RowsAffected()
				if rows > 0 {
					log.Printf("🔄 [Postgres] Updated %s temperature to %s°C", roomName, tempStr)
				}
			}
		}
	}
}

var connectHandler mqtt.OnConnectHandler = func(client mqtt.Client) {
	log.Println("✅ [MQTT] Successfully connected to broker!")

	topic := "home/+/telemetry"
	token := client.Subscribe(topic, 1, messagePubHandler)
	token.Wait()
	if token.Error() != nil {
		log.Printf("❌ [MQTT Error] Failed to subscribe to topic %s: %v", topic, token.Error())
	} else {
		log.Printf("🎧 [MQTT] Subscribed to topic pattern: %s", topic)
	}
}

var connectLostHandler mqtt.ConnectionLostHandler = func(client mqtt.Client, err error) {
	log.Printf("⚠️ [MQTT Warning] Connection lost: %v", err)
}

// InitMQTT принимает параметры для Mongo и Postgres
func InitMQTT(coll *mongo.Collection, db *sql.DB) mqtt.Client {
	mongoCollection = coll
	postgresDB = db

	brokerURL := os.Getenv("MQTT_BROKER")
	if brokerURL == "" {
		brokerURL = os.Getenv("MQTT_BROKER_URL")
	}
	if brokerURL == "" {
		brokerURL = "tcp://broker:1883"
	}

	opts := mqtt.NewClientOptions()
	opts.AddBroker(brokerURL)

	clientID := fmt.Sprintf("smart_home_go_engine_%d", time.Now().UnixNano())
	opts.SetClientID(clientID)

	if user := os.Getenv("MQTT_USER"); user != "" {
		opts.SetUsername(user)
	}
	if pass := os.Getenv("MQTT_PASSWORD"); pass != "" {
		opts.SetPassword(pass)
	}

	opts.SetOnConnectHandler(connectHandler)
	opts.SetConnectionLostHandler(connectLostHandler)
	opts.SetKeepAlive(60 * time.Second)
	opts.SetPingTimeout(10 * time.Second)
	opts.SetAutoReconnect(true)

	client := mqtt.NewClient(opts)
	if token := client.Connect(); token.Wait() && token.Error() != nil {
		log.Printf("❌ [MQTT Error] Failed to connect: %v", token.Error())
	}

	return client
}