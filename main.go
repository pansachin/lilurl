package main

import (
	"fmt"
	"log/slog"
	"os"

	"github.com/gofiber/fiber/v3"
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

	logger.Info("registering routes")
	// Register routes
	routes.RegisterRoutes(app, db, logger)

	logger.Info("starting server")
	app.Listen(":"+cfg.App.Port, fiber.ListenConfig{
		// EnablePrefork:     true,
		EnablePrintRoutes: cfg.Log.PrintRoutes,
	})

	return nil
}
