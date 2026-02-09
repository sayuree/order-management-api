package database

import (
	"database/sql"
	"fmt"
	"log"
	"time"

	_ "github.com/lib/pq"
	"github.com/sabina/orders-api/internal/config"
)

type Database struct {
	DB *sql.DB
}

func New(cfg *config.DatabaseConfig) (*Database, error) {
       // Try to connect to the target database
       db, err := sql.Open("postgres", cfg.ConnectionString())
       if err == nil && db.Ping() == nil {
	       db.SetMaxOpenConns(25)
	       db.SetMaxIdleConns(5)
	       db.SetConnMaxLifetime(5 * time.Minute)
	       log.Println("Database connection established successfully")
	       return &Database{DB: db}, nil
       }

       // If failed, try to create the database
       adminConnStr := fmt.Sprintf("host=%s port=%s user=%s password=%s sslmode=%s dbname=postgres", cfg.Host, cfg.Port, cfg.User, cfg.Password, cfg.SSLMode)
       adminDB, err := sql.Open("postgres", adminConnStr)
       if err != nil {
	       return nil, fmt.Errorf("failed to open admin connection: %w", err)
       }
       defer adminDB.Close()

       // Check if database exists
       var exists bool
       err = adminDB.QueryRow("SELECT 1 FROM pg_database WHERE datname = $1", cfg.DBName).Scan(&exists)
       if err != nil {
	       // If not found, create it
	       _, err = adminDB.Exec("CREATE DATABASE " + cfg.DBName)
	       if err != nil {
		       return nil, fmt.Errorf("failed to create database: %w", err)
	       }
	       log.Printf("Database '%s' created successfully", cfg.DBName)
       }

       // Now connect to the new database
       db, err = sql.Open("postgres", cfg.ConnectionString())
       if err != nil {
	       return nil, fmt.Errorf("failed to open database after creation: %w", err)
       }
       db.SetMaxOpenConns(25)
       db.SetMaxIdleConns(5)
       db.SetConnMaxLifetime(5 * time.Minute)
       if err := db.Ping(); err != nil {
	       return nil, fmt.Errorf("failed to ping database after creation: %w", err)
       }
       log.Println("Database connection established successfully")
       return &Database{DB: db}, nil
}

func (d *Database) Close() error {
	if d.DB != nil {
		return d.DB.Close()
	}
	return nil
}

func (d *Database) Health() error {
	return d.DB.Ping()
}
