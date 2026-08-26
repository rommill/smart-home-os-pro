package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"smart-home/api"
	"smart-home/db"
	"smart-home/mqtt"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("ℹ️ No .env file found, using system environment variables")
	}

	postgresURL := os.Getenv("POSTGRES_URL")
	if postgresURL == "" {
		postgresURL = "postgresql://admin:smart_password@db_relational:5432/smart_home_db?sslmode=disable"
	}

	mongoURI := os.Getenv("MONGO_URI")
	if mongoURI == "" {
		mongoURI = "mongodb://db_nosql:27017"
	}

	pgDB := db.InitPostgres(postgresURL)
	defer pgDB.Close()

	mongoClient := db.InitMongo(mongoURI)
	defer func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := mongoClient.Disconnect(ctx); err != nil {
			log.Printf("⚠️ Error disconnecting MongoDB: %v", err)
		}
	}()

	mongoColl := mongoClient.Database("smart_home").Collection("telemetry")

	// Передаем mongoColl в инициализацию MQTT
	_ = mqtt.InitMQTT(mongoColl)

	r := gin.Default()
	r.Use(api.CORSMiddleware())

	// Public routes
	r.POST("/login", api.Login(pgDB))
	r.POST("/api/v1/telemetry", api.ReceiveTelemetry(mongoColl))

	// Protected routes
	protected := r.Group("/")
	protected.Use(api.AuthMiddleware())
	{
		protected.GET("/telemetry", api.GetTelemetry(pgDB, mongoColl))
		protected.POST("/rooms/:id/target-temp", api.UpdateTargetTemperature(pgDB))
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	port = strings.TrimPrefix(port, ":")

	srv := &http.Server{
		Addr:    ":" + port,
		Handler: r,
	}

	go func() {
		log.Printf("🚀 Server running on http://localhost:%s", port)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("❌ Listen error: %s\n", err)
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("⚠️ Shutting down server gracefully...")

	// The context is used to inform the server it has 5 seconds to finish current requests
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("❌ Server forced to shutdown: %v", err)
	}

	log.Println("✅ Server exited cleanly")
}