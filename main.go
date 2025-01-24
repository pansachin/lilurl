package main

import (
	"log/slog"
	"os"

	"github.com/gofiber/fiber/v3"
	routes "github.com/pansachin/lilurl/app/handlers"
	sqlitedb "github.com/pansachin/lilurl/pkg/database/sqlite"
	"github.com/pansachin/lilurl/pkg/log"
)

func main() {
	logger := log.NewDevelopmentLogger(os.Stdout)
	if err := run(logger); err != nil {
		panic(err)
	}
}

func run(logger *slog.Logger) error {
	// Initialize the SQLite database
	logger.Info("initializing SQLite database")
	db, err := sqlitedb.Initialize()
	if err != nil {
		return err
	}
	defer db.Close()

	logger.Info("initializing fiber app")
	// Initialize the Fiber app
	app := fiber.New(fiber.Config{
		AppName: "LilURL",
	})

	logger.Info("registering routes")
	// Register routes
	routes.RegisterRoutes(app, db)

	logger.Info("starting server")
	app.Listen(":3000", fiber.ListenConfig{
		// EnablePrefork:     true,
		EnablePrintRoutes: true,
	})

	return nil
}
