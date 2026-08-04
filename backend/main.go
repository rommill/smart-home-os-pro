package main

import (
	"fmt"
	"os"
	"smart-home/api"
	"smart-home/db"
	"smart-home/mqtt"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {
	// Load environment variables from .env file if available
	godotenv.Load()

	// Initialize Relational (PostgreSQL) and NoSQL (MongoDB) connections
	pgDB := db.InitPostgres(os.Getenv("POSTGRES_URL"))
	mongoClient := db.InitMongo(os.Getenv("MONGO_URI"))
	mongoColl := mongoClient.Database("smart_home").Collection("telemetry")

	// Apply quick database migrations
	_, migrationErr := pgDB.Exec(`ALTER TABLE rooms ADD COLUMN IF NOT EXISTS target_temperature DOUBLE PRECISION DEFAULT 23.0;`)
	if migrationErr != nil {
		fmt.Println("⚠️ Migration warning/error:", migrationErr)
	} else {
		fmt.Println("✅ PostgreSQL structure verified successfully (target_temperature column active)")
	}

	// Initialize MQTT client subscriber for live telemetry ingestion
	_ = mqtt.InitMQTT(mongoColl)

	// Initialize Gin HTTP router
	r := gin.Default()

	// Global CORS middleware
	r.Use(api.CORSMiddleware())

	// Public routes
	r.POST("/login", api.Login(pgDB))

	// Protected routes (JWT authentication required)
	protected := r.Group("/")
	protected.Use(api.AuthMiddleware())
	{
		protected.GET("/telemetry", api.GetTelemetry(pgDB, mongoColl))
		protected.POST("/rooms/:id/target-temp", api.UpdateTargetTemperature(pgDB))
	}

	// Start server on designated port
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	r.Run(":" + port)
}