package main

import (
	"context"
	"database/sql"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"smart-home/api"
	"smart-home/mqtt"

	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq" // Драйвер PostgreSQL для sql.Open
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
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

	// RU: Надежное подключение к PostgreSQL с циклом повторных попыток (Retry Loop)
	// EN: Reliable PostgreSQL connection with a retry loop
	var pgDB *sql.DB
	var err error
	log.Println("⏳ Connecting to PostgreSQL...")
	for i := 1; i <= 15; i++ {
		pgDB, err = sql.Open("postgres", postgresURL)
		if err == nil {
			ctxPing, cancelPing := context.WithTimeout(context.Background(), 2*time.Second)
			err = pgDB.PingContext(ctxPing)
			cancelPing()
			if err == nil {
				log.Println("✅ PostgreSQL connected successfully!")
				break
			}
		}

		log.Printf("⚠️ PostgreSQL is starting up (attempt %d/15): %v. Retrying in 2s...", i, err)
		if i == 15 {
			log.Fatalf("❌ FATAL: Could not connect to PostgreSQL after 15 attempts: %v", err)
		}
		time.Sleep(2 * time.Second)
	}

	// RU: Надежное подключение к MongoDB с циклом повторных попыток
	// EN: Reliable MongoDB connection with a retry loop
	var mongoClient *mongo.Client
	log.Println("⏳ Connecting to MongoDB...")
	for i := 1; i <= 15; i++ {
		ctxConn, cancelConn := context.WithTimeout(context.Background(), 5*time.Second)
		mongoClient, err = mongo.Connect(ctxConn, options.Client().ApplyURI(mongoURI))
		cancelConn()

		if err == nil {
			ctxPing, cancelPing := context.WithTimeout(context.Background(), 2*time.Second)
			err = mongoClient.Ping(ctxPing, nil)
			cancelPing()
			if err == nil {
				log.Println("✅ MongoDB connected successfully!")
				break
			}
		}

		log.Printf("⚠️ MongoDB is starting up (attempt %d/15): %v. Retrying in 2s...", i, err)
		if i == 15 {
			log.Fatalf("❌ FATAL: Could not connect to MongoDB after 15 attempts: %v", err)
		}
		time.Sleep(2 * time.Second)
	}

	mongoColl := mongoClient.Database("smart_home").Collection("telemetry")

	// RU: Создаем и запускаем WebSocket Hub для рассылки данных в реальном времени
	// EN: Create and start the WebSocket Hub for real-time data broadcasting
	wsHub := api.NewHub()
	go wsHub.Run()

	// RU: Инициализируем MQTT подписчик и передаем wsHub для рассылки сообщений
	// EN: Initialize MQTT subscriber and pass wsHub for message broadcasting
	_ = mqtt.InitMQTT(mongoColl, pgDB, wsHub)

	r := gin.Default()
	r.Use(api.PrometheusMiddleware())
	r.Use(api.CORSMiddleware())

	// --- Healthcheck & Readiness Probes ---

	// RU: Liveness probe: проверяет, что HTTP-сервер работает
	// EN: Liveness probe: verifies the HTTP server is running
	r.GET("/healthz", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "UP"})
	})

	// RU: Readiness probe: проверяет доступность баз данных перед приемом трафика
	// EN: Readiness probe: verifies DB connectivity before serving traffic
	r.GET("/readyz", func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(c.Request.Context(), 2*time.Second)
		defer cancel()

		// RU: Проверка подключения к PostgreSQL
		// EN: Check PostgreSQL connection health
		if err := pgDB.PingContext(ctx); err != nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{
				"status": "DOWN",
				"error":  "PostgreSQL unavailable: " + err.Error(),
			})
			return
		}

		// RU: Проверка подключения к MongoDB
		// EN: Check MongoDB connection health
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

	// --- WebSocket Endpoint ---
	// RU: Маршрут для подключения фронтенд-клиентов по WebSocket
	// EN: Endpoint for frontend WebSocket connections
	r.GET("/ws", func(c *gin.Context) {
		api.ServeWS(wsHub, c.Writer, c.Request)
	})

	// --- Public Endpoints ---
	r.POST("/login", api.LoginRateLimiter(5, 1*time.Minute), api.Login(pgDB))
	r.POST("/api/v1/telemetry", api.ReceiveTelemetry(mongoColl))

	// --- Service / Emulator Endpoints (Protected by API Key) ---
	// RU: Эндпоинты для фоновых сервисов и эмуляторов, защищенные системным API-ключом
	// EN: Endpoints for background services and emulators, protected by a system API key
	serviceGroup := r.Group("/api/v1/service")
	serviceGroup.Use(api.ServiceAuthMiddleware())
	{
		serviceGroup.GET("/targets", api.GetTelemetry(pgDB, mongoColl))
	}

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

	// RU: Конфигурация HTTP-сервера с таймаутами для продакшена
	// EN: Configure HTTP server with production-ready timeouts
	srv := &http.Server{
		Addr:         ":" + port,
		Handler:      r,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	// RU: Запуск HTTP-сервера в отдельной горутине
	// EN: Start HTTP server in a non-blocking goroutine
	go func() {
		log.Printf("🚀 Server running on http://localhost:%s", port)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("❌ Server ListenAndServe error: %s\n", err)
		}
	}()

	// RU: Перехват сигналов завершения от операционной системы (SIGINT, SIGTERM)
	// EN: Intercept OS shutdown signals (SIGINT, SIGTERM)
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("⚠️ Shutting down server gracefully...")

	// RU: Таймаут 5 секунд для завершения текущих HTTP-запросов
	// EN: Context timeout for in-flight HTTP request completion (5 seconds)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("❌ Server forced to shutdown: %v", err)
	}

	// RU: Безопасное закрытие пулов соединений с базами данных
	// EN: Safely close database connection pools after HTTP server shutdown
	log.Println("🔌 Closing database connections...")
	pgDB.Close()

	mongoCtx, mongoCancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer mongoCancel()
	if err := mongoClient.Disconnect(mongoCtx); err != nil {
		log.Printf("⚠️ Error disconnecting MongoDB: %v", err)
	}

	log.Println("✅ Server exited cleanly")
}