package main

import (
	"context"
	"database/sql"
	"flag"
	"log"
	"os"
	"strconv"
	"time"

	"biling_api/internal/data"

	"github.com/joho/godotenv"

	_ "github.com/lib/pq"
)

const version = "1.0.0"

// config holds application configuration
type config struct {
	port int
	env  string
	db   struct {
		dsn          string
		maxOpenConns int
		maxIdleConns int
		maxIdleTime  string
	}
	jwt struct {
		secret string
	}
}

// application holds dependencies
type application struct {
	config config
	logger *log.Logger
	models data.Models
}

func init() {
	err := godotenv.Load(".env")

	if err != nil {
		log.Fatal("Error loading .env file")
	}
}

func main() {
	var cfg config

	// Parse command-line flags
	flag.IntVar(&cfg.port, "port", getIntEnv("PORT", 4000), "API server port")
	flag.StringVar(&cfg.env, "env", getEnv("ENV", "development"), "Environment (development|staging|production)")
	flag.StringVar(&cfg.db.dsn, "db-dsn", getEnv("DB_DSN", ""), "PostgreSQL DSN")
	flag.IntVar(&cfg.db.maxOpenConns, "db-max-open-conns", getIntEnv("DB_MAX_OPEN_CONNS", 25), "PostgreSQL max open connections")
	flag.IntVar(&cfg.db.maxIdleConns, "db-max-idle-conns", getIntEnv("DB_MAX_IDLE_CONNS", 25), "PostgreSQL max idle connections")
	flag.StringVar(&cfg.db.maxIdleTime, "db-max-idle-time", getEnv("DB_MAX_IDLE_TIME", "15m"), "PostgreSQL max connection idle time")
	flag.StringVar(&cfg.jwt.secret, "jwt-secret", getEnv("JWT_SECRET", ""), "JWT secret key")
	flag.Parse()

	// Create logger
	logger := log.New(os.Stdout, "", log.Ldate|log.Ltime)

	// Open database connection
	db, err := openDB(cfg)
	if err != nil {
		logger.Fatal(err)
	}
	defer db.Close()

	logger.Printf("database connection pool established")

	// Initialize application
	app := &application{
		config: cfg,
		logger: logger,
		models: data.NewModels(db),
	}

	// Set JWT secret in token model
	app.models.Tokens.Secret = cfg.jwt.secret

	// Start server
	err = app.serve()
	if err != nil {
		logger.Fatal(err)
	}
}

func getEnv(env string, value string) string {
	if v := os.Getenv(env); v != "" {
		return v
	}
	return value
}

func getIntEnv(env string, value int) int {
	v := os.Getenv(env)
	if v == "" {
		return value
	}
	n, err := strconv.Atoi(v)
	if err != nil {
		log.Printf("warning: failed to parse %s as integer, using default value %d", env, value)
		return value
	}
	return n
}

func openDB(cfg config) (*sql.DB, error) {
	db, err := sql.Open("postgres", cfg.db.dsn)
	if err != nil {
		return nil, err
	}

	db.SetMaxOpenConns(cfg.db.maxOpenConns)
	db.SetMaxIdleConns(cfg.db.maxIdleConns)

	duration, err := time.ParseDuration(cfg.db.maxIdleTime)
	if err != nil {
		return nil, err
	}

	db.SetConnMaxIdleTime(duration)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err = db.PingContext(ctx)
	if err != nil {
		return nil, err
	}

	return db, nil
}
