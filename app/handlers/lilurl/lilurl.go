package lilurl

import (
	"log/slog"
	"strconv"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/jmoiron/sqlx"
	model "github.com/pansachin/lilurl/app/models/lilurl"
)

type Handler struct {
	db     *model.Core
	logger *slog.Logger
}

// NewHandler initializes a new handler
func NewHandler(db *sqlx.DB, logger *slog.Logger) *Handler {
	return &Handler{
		db:     model.New(db, logger),
		logger: logger,
	}
}

func (h *Handler) Create(c fiber.Ctx) error {
	var payload model.CreateLilURL

	if err := c.Bind().Body(&payload); err != nil {
		return c.SendStatus(fiber.StatusBadRequest)
	}

	now := time.Now().Truncate(time.Second)
	payload.CretedAt = now
	payload.UpdatedAt = now
	res, err := h.db.Create(payload)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}
	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"result": res,
	})
}

// GetByID retrieves a lilurl by its ID
func (h *Handler) GetByID(c fiber.Ctx) error {
	id := c.Params("id")
	if id == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "id is required",
		})
	}

	intId, err := strconv.Atoi(id)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "id must be a number",
		})
	}
	int64Id := int64(intId)

	res, err := h.db.GetByID(int64Id)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"result": res,
	})
}

// GetByShortURL retrieves a lilurl by its ID
func (h *Handler) GetByShortURL(c fiber.Ctx) error {
	lilurl := c.Params("lilurl")
	if lilurl == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "hash is required",
		})
	}

	res, err := h.db.GetByShortURL(lilurl)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"result": res,
	})
}

// Get return the original URL by lilurl
func (h *Handler) Get(c fiber.Ctx) error {
	lilurl := c.Params("lilurl")
	if lilurl == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "hash is required",
		})
	}

	res, err := h.db.GetByShortURL(lilurl)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	c.Response().Header.Add("Location", res.Long)
	return c.Status(fiber.StatusTemporaryRedirect).Send(nil)
}
