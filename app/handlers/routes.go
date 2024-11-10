package app

import (
	"github.com/gofiber/fiber/v3"
	"github.com/jmoiron/sqlx"
	"github.com/pansachin/lilurl/app/handlers/lilurl"
)

func RegisterRoutes(app *fiber.App, db *sqlx.DB) {
	// Initialize the handler
	h := lilurl.NewHandler(db)

	// Get a url by short url
	app.Get("/:lilurl", h.GetByShortURL)

	// Get a details by id
	app.Get("/v1/:id", h.GetByID)

	// Create a new short url
	app.Post("/v1/lilurl", h.Post)
}
