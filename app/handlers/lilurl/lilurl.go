package lilurl

import (
	"database/sql"

	"github.com/gofiber/fiber/v3"
)

type Handler struct {
	db *sql.DB
}

// NewHandler initializes a new handler
func NewHandler(db *sql.DB) *Handler {
	return &Handler{
		db: db,
	}
}

func (h *Handler) Get(c fiber.Ctx) error {
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "Hello, World!",
	})
}

func (h *Handler) Post(c fiber.Ctx) error {
	return c.SendString("Hello, World!")
}
