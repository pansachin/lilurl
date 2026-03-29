package lilurl

import (
	"errors"
	"log/slog"
	"strconv"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/jmoiron/sqlx"
	model "github.com/pansachin/lilurl/app/models/lilurl"
	store "github.com/pansachin/lilurl/app/models/lilurl/db"
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
	payload.CreatedAt = now
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

	res, err := h.db.GetByID(int64(intId))
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	if res.ID == 0 {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "not found",
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"result": res,
	})
}

// GetByShortURL retrieves a lilurl by its short URL
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

	if res.ID == 0 {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "not found",
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"result": res,
	})
}

// Delete soft-deletes a lilurl by its ID
func (h *Handler) Delete(c fiber.Ctx) error {
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

	if err := h.db.Delete(int64(intId)); err != nil {
		if errors.Is(err, store.ErrNotFound) {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error": "not found",
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.SendStatus(fiber.StatusNoContent)
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

	if res.ID == 0 {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "not found",
		})
	}

	c.Response().Header.Add("Location", res.Long)
	return c.Status(fiber.StatusTemporaryRedirect).Send(nil)
}
