package main

import (
	"fmt" // Добавился импорт fmt для логов миграции
	"os"
	"smart-home/api"
	"smart-home/db"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {
	godotenv.Load()
	pgDB := db.InitPostgres(os.Getenv("POSTGRES_URL"))
	mongoClient := db.InitMongo(os.Getenv("MONGO_URI"))
	mongoColl := mongoClient.Database("smart_home").Collection("telemetry")

	// --- АВТО-МИГРАЦИЯ БАЗЫ ДАННЫХ (НАШЕ ИСПРАВЛЕНИЕ) ---
	// Этот код сам добавит колонку target_temperature в PostgreSQL, если её там ещё нет!
	_, migrationErr := pgDB.Exec(`ALTER TABLE rooms ADD COLUMN IF NOT EXISTS target_temperature DOUBLE PRECISION DEFAULT 23.0;`)
	if migrationErr != nil {
		fmt.Println("⚠️ Migration warning/error:", migrationErr)
	} else {
		fmt.Println("✅ PostgreSQL structure verified successfully (target_temperature column active)")
	}
	// ----------------------------------------------------

	r := gin.Default()

	r.Use(api.CORSMiddleware())

	r.POST("/login", api.Login(pgDB))

	protected := r.Group("/")
	protected.Use(api.AuthMiddleware())
	{
		protected.GET("/telemetry", api.GetTelemetry(pgDB, mongoColl))
		protected.POST("/rooms/:id/target-temp", api.UpdateTargetTemperature(pgDB))
	}

	r.Run(":8080")
}