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
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func main() {
	jwtSecret := os.Getenv("JWT_SECRET")
if jwtSecret == "" || jwtSecret == "change_me_in_production" {
    log.Fatal("FATAL: JWT_SECRET is not set or using insecure default value!")
}

	postgresURL := os.Getenv("POSTGRES_URL")
	if postgresURL == "" {
		postgresURL = "postgresql://admin:smart_password@db_relational:5432/smart_home_db?sslmode=disable"
	}

	mongoURI := os.Getenv("MONGO_URI")
	if mongoURI == "" {
		mongoURI = "mongodb://db_nosql:27017"
	}

	// Initialize database connection pools
	pgDB := db.InitPostgres(postgresURL)
	mongoClient := db.InitMongo(mongoURI)
	mongoColl := mongoClient.Database("smart_home").Collection("telemetry")

	// Initialize MQTT client subscriber for real-time sensor ingestion
	_ = mqtt.InitMQTT(mongoColl, pgDB)

	r := gin.Default()
r.Use(api.PrometheusMiddleware())
r.Use(api.CORSMiddleware())

	// --- Healthcheck & Readiness Probes ---

	// Liveness probe: verifies the HTTP server is running
	r.GET("/healthz", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "UP"})
	})

	// Readiness probe: verifies DB connectivity before serving traffic
	r.GET("/readyz", func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(c.Request.Context(), 2*time.Second)
		defer cancel()

		// Check PostgreSQL connection health
		if err := pgDB.PingContext(ctx); err != nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{
				"status": "DOWN",
				"error":  "PostgreSQL unavailable: " + err.Error(),
			})
			return
		}

		// Check MongoDB connection health
		if err := mongoClient.Ping(ctx, nil); err != nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{
				"status": "DOWN",
				"error":  "MongoDB unavailable: " + err.Error(),
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"status":   "READY",
			"postgres": "connected",
			"mongodb":  "connected",
		})
	})

	// --- Prometheus Metrics Endpoint ---
	r.GET("/metrics", gin.WrapH(promhttp.Handler()))

	// --- Public Endpoints ---
	r.POST("/login", api.LoginRateLimiter(5, 1*time.Minute), api.Login(pgDB))
	r.POST("/api/v1/telemetry", api.ReceiveTelemetry(mongoColl))

	// --- Protected Endpoints (Requires JWT Authentication) ---
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

	// Configure HTTP server with production-ready timeouts
	srv := &http.Server{
		Addr:         ":" + port,
		Handler:      r,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	// Start HTTP server in a non-blocking goroutine
	go func() {
		log.Printf("🚀 Server running on http://localhost:%s", port)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("❌ Server ListenAndServe error: %s\n", err)
		}
	}()

	// Intercept OS shutdown signals (SIGINT, SIGTERM)
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("⚠️ Shutting down server gracefully...")

	// Context timeout for in-flight HTTP request completion (5 seconds)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("❌ Server forced to shutdown: %v", err)
	}

	// Safely close database connection pools after HTTP server shutdown
	log.Println("🔌 Closing database connections...")
	pgDB.Close()

	mongoCtx, mongoCancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer mongoCancel()
	if err := mongoClient.Disconnect(mongoCtx); err != nil {
		log.Printf("⚠️ Error disconnecting MongoDB: %v", err)
	}

	log.Println("✅ Server exited cleanly")
}