package app

import (
	"database/sql"

	"github.com/gofiber/fiber/v3"
	"github.com/pansachin/lilurl/app/handlers/lilurl"
)

func RegisterRoutes(app *fiber.App, db *sql.DB) {
	// Initialize the handler
	h := lilurl.NewHandler(db)

	//
	app.Get("/", h.Get)

	//
	app.Post("/v1/lilurl", h.Post)
}
