package app

import (
	"log/slog"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/limiter"
	"github.com/jmoiron/sqlx"
	"github.com/pansachin/lilurl/app/handlers/lilurl"
	"github.com/pansachin/lilurl/config"
)

// Register routes
func RegisterRoutes(app *fiber.App, db *sqlx.DB, log *slog.Logger, rl *config.RateLimit) {
	// Initialize the handler
	h := lilurl.NewHandler(db, log)

	// Health check
	app.Get("/health", func(c fiber.Ctx) error {
		return c.Status(fiber.StatusOK).JSON(fiber.Map{
			"status": "ok",
		})
	})

	// URL redirection
	app.Get("/:lilurl", h.Get)

	// Get details by id (numeric) or short url (alphanumeric)
	app.Get("/api/v1/:param", h.GetByParam)

	// Create a new short url with stricter per-route rate limit.
	// Handlers execute left-to-right: limiter runs first, then the handler.
	// Uses a "create:" key prefix to keep its counter separate from the global limiter.
	// Configurable via rate_limit.create_max and rate_limit.create_window_secs in config.yaml.
	createLimiter := limiter.New(limiter.Config{
		Max:        rl.CreateMax,
		Expiration: time.Duration(rl.CreateWindowSecs) * time.Second,
		KeyGenerator: func(c fiber.Ctx) string {
			return "create:" + c.IP()
		},
		LimiterMiddleware: limiter.SlidingWindow{},
	})
	app.Post("/api/v1/lilurl", createLimiter, h.Create)

	// Delete a short url by id (soft delete)
	app.Delete("/api/v1/:id", h.Delete)
}
