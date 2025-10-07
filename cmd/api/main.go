package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-historical-data/internal/controller"
	"github.com/go-historical-data/internal/middleware"
	"github.com/go-historical-data/internal/model"
	"github.com/go-historical-data/internal/repository"
	"github.com/go-historical-data/internal/service"
	"github.com/go-historical-data/pkg/config"
	"github.com/go-historical-data/pkg/database"
	applogger "github.com/go-historical-data/pkg/logger"
	"github.com/go-historical-data/pkg/validator"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/compress"
)

func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		fmt.Printf("Failed to load configuration: %v\n", err)
		os.Exit(1)
	}

	// Initialize logger
	log := applogger.New(applogger.Config{
		Level:  cfg.Logging.Level,
		Format: cfg.Logging.Format,
	})

	log.Info().Msg("Starting Historical Data API...")

	// Connect to MySQL
	dbLogLevel := database.GetLogLevel(cfg.Logging.Level)
	db, err := database.NewMySQLConnection(cfg.Database, dbLogLevel)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to connect to database")
	}
	log.Info().Msg("Connected to MySQL database")

	// Auto-migrate database schema
	if err := db.AutoMigrate(&model.HistoricalData{}); err != nil {
		log.Fatal().Err(err).Msg("Failed to migrate database schema")
	}
	log.Info().Msg("Database schema migrated successfully")

	// Initialize validator
	v := validator.New()

	// Initialize repository
	historicalRepo := repository.NewHistoricalRepository(db)

	// Initialize service
	historicalService := service.NewHistoricalService(historicalRepo)

	// Initialize controllers
	healthController := controller.NewHealthController()
	historicalController := controller.NewHistoricalController(historicalService, v)

	// Initialize Fiber app
	app := fiber.New(fiber.Config{
		ErrorHandler:          middleware.ErrorHandler(),
		DisableStartupMessage: true,
		AppName:               cfg.App.Name,
		ReadTimeout:           time.Duration(cfg.API.RequestTimeout) * time.Second,
		WriteTimeout:          time.Duration(cfg.API.RequestTimeout) * time.Second,
	})

	// Global middleware
	app.Use(middleware.Recover())
	app.Use(middleware.RequestID())
	app.Use(middleware.Logger(log))
	app.Use(middleware.CORS(cfg.CORS))
	app.Use(compress.New(compress.Config{
		Level: compress.LevelBestSpeed,
	}))

	// Rate limiting
	if cfg.API.RateLimit > 0 {
		app.Use(middleware.RateLimiter(cfg.API.RateLimit))
	}

	// Health check routes
	app.Get("/health", healthController.Check)

	// API routes
	apiV1 := app.Group("/api/v1")
	{
		// Historical data endpoints
		apiV1.Post("/data", historicalController.UploadCSV)
		apiV1.Get("/data", historicalController.GetData)
		apiV1.Get("/data/:id", historicalController.GetDataByID)
	}

	// Start server in a goroutine
	go func() {
		addr := fmt.Sprintf(":%d", cfg.App.Port)
		log.Info().
			Str("address", addr).
			Str("env", cfg.App.Env).
			Msg("Server starting")

		if err := app.Listen(addr); err != nil {
			log.Fatal().Err(err).Msg("Failed to start server")
		}
	}()

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	<-quit

	log.Info().Msg("Shutting down server...")

	// Shutdown with timeout
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(cfg.API.ShutdownTimeout)*time.Second)
	defer cancel()

	if err := app.ShutdownWithContext(ctx); err != nil {
		log.Error().Err(err).Msg("Server forced to shutdown")
	}

	// Close database connections
	sqlDB, _ := db.DB()
	if sqlDB != nil {
		sqlDB.Close()
	}

	log.Info().Msg("Server exited gracefully")
}
