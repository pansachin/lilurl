package main

import (
	"github.com/gofiber/fiber/v3"
	routes "github.com/pansachin/lilurl/app/handlers"
	sqlitedb "github.com/pansachin/lilurl/pkg/database/sqlite"
)

func main() {
	if err := run(); err != nil {
		panic(err)
	}
}

func run() error {
	// Initialize the SQLite database
	db, err := sqlitedb.Initialize()
	if err != nil {
		return err
	}

	// Initialize the Fiber app
	app := fiber.New(fiber.Config{
		AppName: "LilURL",
	})

	// Register routes
	routes.RegisterRoutes(app, db)

	app.Listen(":3000")
	return nil
}
