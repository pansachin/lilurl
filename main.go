package main

import (
	"fmt"
	"log/slog"
	"os"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/limiter"
	sqlitedb "github.com/pansachin/lilurl/pkg/database/sqlite"

	routes "github.com/pansachin/lilurl/app/handlers"
	"github.com/pansachin/lilurl/config"
	"github.com/pansachin/lilurl/pkg/log"
)

func main() {
	logger := log.NewProductionLogger(os.Stdout)
	if err := run(logger); err != nil {
		panic(err)
	}
}

func run(logger *slog.Logger) error {
	// Read configuration
	cfg := new(config.Config)
	if err := config.Read(cfg, "config.yaml", "/config/config.yaml"); err != nil {
		return err
	}

	// Log configuration
	if cfg.Log.Debug {
		logger = log.NewDevelopmentLogger(os.Stdout)
		logger.Debug("configuration", "cfg", fmt.Sprintf("%#v", cfg))
	}

	// Initialize the SQLite database
	logger.Info("initializing SQLite database")
	db, err := sqlitedb.Initialize(cfg)
	if err != nil {
		return err
	}
	defer db.Close()

	logger.Info("initializing fiber app")
	// Initialize the Fiber app
	app := fiber.New(fiber.Config{
		AppName: cfg.App.Name,
	})

	// Configure CORS middleware
	app.Use(func(c fiber.Ctx) error {
		c.Set("Access-Control-Allow-Origin", cfg.App.CORSOrigin)
		c.Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		c.Set("Access-Control-Allow-Credentials", "true")

		// Handle preflight requests
		if c.Method() == "OPTIONS" {
			return c.SendStatus(fiber.StatusNoContent)
		}

		return c.Next()
	})

	// Global rate limiter — applies to all endpoints (GET redirects, API reads, etc.).
	// Uses sliding window algorithm with in-memory storage, keyed by client IP.
	// Configurable via rate_limit.max and rate_limit.window_secs in config.yaml.
	app.Use(limiter.New(limiter.Config{
		Max:               cfg.RateLimit.Max,
		Expiration:        time.Duration(cfg.RateLimit.WindowSecs) * time.Second,
		LimiterMiddleware: limiter.SlidingWindow{},
	}))

	logger.Info("registering routes")
	// Register routes
	routes.RegisterRoutes(app, db, logger, &cfg.RateLimit)

	logger.Info("starting server")
	app.Listen(":"+cfg.App.Port, fiber.ListenConfig{
		// EnablePrefork:     true,
		EnablePrintRoutes: cfg.Log.PrintRoutes,
	})

	return nil
}
