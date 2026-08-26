package db

import (
	"context"
	"database/sql"
	"log"

	_ "github.com/lib/pq"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"golang.org/x/crypto/bcrypt"
)

func InitPostgres(connStr string) *sql.DB {
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		log.Fatal("❌ [PG ERROR] Connection failed:", err)
	}

	if err := db.Ping(); err != nil {
		log.Fatal("❌ [PG ERROR] Ping failed:", err)
	}

	// 1. Создание таблиц и гарантированное добавление колонки role
	createTablesSQL := `
	CREATE TABLE IF NOT EXISTS users (
		id SERIAL PRIMARY KEY,
		username VARCHAR(50) UNIQUE NOT NULL,
		password_hash TEXT NOT NULL,
		role VARCHAR(20) DEFAULT 'admin'
	);
	ALTER TABLE users ADD COLUMN IF NOT EXISTS role VARCHAR(20) DEFAULT 'admin';

	CREATE TABLE IF NOT EXISTS rooms (
		id SERIAL PRIMARY KEY,
		name VARCHAR(100) NOT NULL,
		target_temperature NUMERIC(4,2) DEFAULT 22.0
	);`

	if _, err := db.Exec(createTablesSQL); err != nil {
		log.Printf("⚠️ [PG MIGRATION NOTICE] %v", err)
	}

	// 2. Сидинг / обнуление пароля для admin (логин: admin, пароль: admin)
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte("admin"), bcrypt.DefaultCost)
	if err == nil {
		_, err = db.Exec(`
			INSERT INTO users (username, password_hash, role) 
			VALUES ($1, $2, $3)
			ON CONFLICT (username) 
			DO UPDATE SET password_hash = EXCLUDED.password_hash, role = EXCLUDED.role;
		`, "admin", string(hashedPassword), "admin")

		if err != nil {
			log.Printf("⚠️ Failed to seed admin user: %v", err)
		} else {
			log.Println("✅ [PG SEED] Default admin user ready (admin / admin)")
		}
	}

	// 3. Сидинг комнат (ровно 2 комнаты)
	_, err = db.Exec(`
        INSERT INTO rooms (id, name, target_temperature) VALUES 
        (1, 'Living Room', 22.5),
        (2, 'Bedroom', 21.0)
        ON CONFLICT (id) DO NOTHING;
    `)
    if err != nil {
        log.Printf("⚠️ Failed to seed rooms: %v", err)
    } else {
        log.Println("✅ [PG SEED] Default rooms check completed")
    }

	return db
}

func InitMongo(uri string) *mongo.Client {
	ctx := context.TODO()
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(uri))
	if err != nil {
		log.Fatal("Ошибка Mongo:", err)
	}
	return client
}