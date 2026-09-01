package db

import (
	"context"
	"database/sql"
	"log"
	"time"

	_ "github.com/lib/pq"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"golang.org/x/crypto/bcrypt"
)

// InitPostgres initializes the PostgreSQL connection, runs initial database migrations, and seeds default records.
func InitPostgres(connStr string) *sql.DB {
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		log.Fatal("❌ [PG ERROR] Database connection initialization failed:", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		log.Fatal("❌ [PG ERROR] Database ping failed:", err)
	}

	// 1. Database schema migration: create essential tables and ensure necessary columns exist
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

	if _, err := db.ExecContext(ctx, createTablesSQL); err != nil {
		log.Printf("⚠️ [PG MIGRATION NOTICE] Table schema setup warning: %v", err)
	}

	// 2. Database seeding: ensure default admin user exists with bcrypt hashed credentials
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte("admin"), bcrypt.DefaultCost)
	if err == nil {
		_, err = db.ExecContext(ctx, `
			INSERT INTO users (username, password_hash, role) 
			VALUES ($1, $2, $3)
			ON CONFLICT (username) 
			DO UPDATE SET password_hash = EXCLUDED.password_hash, role = EXCLUDED.role;
		`, "admin", string(hashedPassword), "admin")

		if err != nil {
			log.Printf("⚠️ [PG SEED WARNING] Failed to seed admin user: %v", err)
		} else {
			log.Println("✅ [PG SEED] Default admin account synchronized (admin / admin)")
		}
	} else {
		log.Printf("❌ [PG SEED ERROR] Failed to hash default admin password: %v", err)
	}

	// 3. Database seeding: ensure default room entries exist
	_, err = db.ExecContext(ctx, `
		INSERT INTO rooms (id, name, target_temperature) VALUES 
		(1, 'Living Room', 22.5),
		(2, 'Bedroom', 21.0)
		ON CONFLICT (id) DO NOTHING;
	`)
	if err != nil {
		log.Printf("⚠️ [PG SEED WARNING] Failed to seed default rooms: %v", err)
	} else {
		log.Println("✅ [PG SEED] Default room records verified")
	}

	return db
}

// InitMongo connects to MongoDB instance using a explicit 10-second connection timeout context.
func InitMongo(uri string) *mongo.Client {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	client, err := mongo.Connect(ctx, options.Client().ApplyURI(uri))
	if err != nil {
		log.Fatal("❌ [MONGO ERROR] Failed to connect to MongoDB instance:", err)
	}

	if err := client.Ping(ctx, nil); err != nil {
		log.Fatal("❌ [MONGO ERROR] MongoDB ping failed:", err)
	}

	log.Println("✅ [MONGO] Connected to MongoDB successfully")
	return client
}