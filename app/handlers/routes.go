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

	// URL redirection
	app.Get("/:lilurl", h.Get)

	// Get details by short url
	app.Get("/api/v1/:lilurl", h.GetByShortURL)

	// Get a details by id
	app.Get("/api/v1/:id", h.GetByID)

	// Create a new short url with stricter per-route rate limit.
	// In Fiber v3, app.Post(path, handler, middleware...) executes middleware
	// before the handler — the variadic middleware args run first in order,
	// then the handler runs last (see router.go register()).
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
	app.Post("/api/v1/lilurl", h.Create, createLimiter)
}
