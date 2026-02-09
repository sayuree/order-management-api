package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/sabina/orders-api/internal/config"
	"github.com/sabina/orders-api/internal/database"
	"github.com/sabina/orders-api/internal/handlers"
	"github.com/sabina/orders-api/internal/repository"
	"github.com/sabina/orders-api/internal/service"
)

func main() {
	// Check for migration commands
		       if len(os.Args) > 1 {
			       switch os.Args[1] {
			       case "migrate":
				       runMigrations(true)
				       return
			       case "migrate-down":
				       runMigrations(false)
				       return
			       case "seed":
				       runSeed()
				       return
			       }
		       }

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Initialize database
	db, err := database.New(&cfg.Database)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	// Initialize repository, service, and handlers
	orderRepo := repository.NewPostgresOrderRepository(db.DB)
	orderService := service.NewOrderService(orderRepo)
	orderHandler := handlers.NewOrderHandler(orderService, cfg.Pagination.DefaultPageSize, cfg.Pagination.MaxPageSize)

	// Setup routes
	router := handlers.SetupRoutes(orderHandler)

	// Create HTTP server
	serverAddr := fmt.Sprintf("%s:%s", cfg.Server.Host, cfg.Server.Port)
	server := &http.Server{
		Addr:         serverAddr,
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Start server in goroutine
	go func() {
		log.Printf("Server starting on %s", serverAddr)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("Server exited")
}

func runSeed() {
       cfg, err := config.Load()
       if err != nil {
	       log.Fatalf("Failed to load configuration: %v", err)
       }

       db, err := database.New(&cfg.Database)
       if err != nil {
	       log.Fatalf("Failed to connect to database: %v", err)
       }
       defer db.Close()

       // Seed 50 sample orders
       if err := database.SeedOrders(db.DB, 50); err != nil {
	       log.Fatalf("Failed to seed orders: %v", err)
       }
       log.Println("Successfully seeded 50 sample orders.")
}

func runMigrations(up bool) {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	db, err := database.New(&cfg.Database)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	// Get absolute path to migrations directory
	wd, err := os.Getwd()
	if err != nil {
		log.Fatalf("Failed to get working directory: %v", err)
	}
	migrationsPath := filepath.Join(wd, "migrations")

	if up {
		if err := database.RunMigrations(db.DB, migrationsPath); err != nil {
			log.Fatalf("Failed to run migrations: %v", err)
		}
	} else {
		if err := database.RollbackMigrations(db.DB, migrationsPath); err != nil {
			log.Fatalf("Failed to rollback migrations: %v", err)
		}
	}
}
