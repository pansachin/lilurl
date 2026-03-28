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

	// Stricter rate limit for URL creation (POST /api/v1/lilurl only).
	// Applied via app.Use() with a Next filter because Fiber v3 beta does not
	// invoke inline per-route middleware (e.g. app.Post(path, mw, handler)).
	// Uses a "create:" key prefix to keep its counter separate from the global limiter.
	// Configurable via rate_limit.create_max and rate_limit.create_window_secs in config.yaml.
	app.Use(limiter.New(limiter.Config{
		Max:        rl.CreateMax,
		Expiration: time.Duration(rl.CreateWindowSecs) * time.Second,
		Next: func(c fiber.Ctx) bool {
			// Skip this limiter for all requests except POST /api/v1/lilurl
			return !(c.Method() == fiber.MethodPost && c.Path() == "/api/v1/lilurl")
		},
		KeyGenerator: func(c fiber.Ctx) string {
			return "create:" + c.IP()
		},
		LimiterMiddleware: limiter.SlidingWindow{},
	}))

	// Create a new short url
	app.Post("/api/v1/lilurl", h.Create)
}
