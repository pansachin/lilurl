package app

import (
	"log/slog"

	"github.com/gofiber/fiber/v3"
	"github.com/jmoiron/sqlx"
	"github.com/pansachin/lilurl/app/handlers/lilurl"
)

// Register routes
func RegisterRoutes(app *fiber.App, db *sqlx.DB, log *slog.Logger) {
	// Initialize the handler
	h := lilurl.NewHandler(db, log)

	// UR redirection
	app.Get("/:lilurl", h.Get)

	// Get details by short url
	app.Get("/api/v1/:lilurl", h.GetByShortURL)

	// Get a details by id
	app.Get("/api/v1/:id", h.GetByID)

	// Create a new short url
	app.Post("/api/v1/lilurl", h.Create)
}
