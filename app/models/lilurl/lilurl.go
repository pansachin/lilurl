package model

import (
	"time"
	"unsafe"

	validate "github.com/go-playground/validator/v10"
	"github.com/jmoiron/sqlx"
	store "github.com/pansachin/lilurl/app/models/lilurl/db"
)

// LilURL model
type LilURL struct {
	ID        int        `json:"id" validate:"required"`
	Long      string     `json:"long_url" validate:"required"`
	Short     string     `json:"short" validate:"required"`
	CreatedAt time.Time  `json:"created_at" validate:"required"`
	UpdatedAt time.Time  `json:"updated_at" validate:"required"`
	DeletedAt *time.Time `json:"deleted_at" validate:"required"`
}

// CreateLilURL model
type CreateLilURL struct {
	Long      string    `json:"long_url" validate:"required"`
	Short     string    `json:"short" validate:"required"`
	CretedAt  time.Time `json:"created_at" validate:"required"`
	UpdatedAt time.Time `json:"updated_at" validate:"required"`
}

type Core struct {
	db *store.Store
}

// New creates a new model
func New(db *sqlx.DB) *Core {
	return &Core{
		db: store.New(db),
	}
}

// Create creates a new lilurl in the database
func (c *Core) Create(payload CreateLilURL) (LilURL, error) {
	if err := validate.New().Struct(payload); err != nil {
		return LilURL{}, err
	}

	data := store.LilURL{
		Long:      payload.Long,
		Short:     payload.Short,
		CreatedAt: payload.CretedAt,
		UpdatedAt: payload.UpdatedAt,
	}

	result, err := c.db.Create(data)
	if err != nil {
		return LilURL{}, err
	}
	return toLilURL(result), nil
}

// GetByID retrieves a lilurl by its ID
func (c *Core) GetByID(id int64) (LilURL, error) {
	result, err := c.db.GetByID(id)
	if err != nil {
		return LilURL{}, err
	}
	return toLilURL(result), nil
}

// GetByShortURL retrieves a lilurl by its short url
func (c *Core) GetByShortURL(short string) (LilURL, error) {
	result, err := c.db.GetByShortURL(short)
	if err != nil {
		return LilURL{}, err
	}
	return toLilURL(result), nil
}

func toLilURL(data store.LilURL) LilURL {
	result := (*LilURL)(unsafe.Pointer(&data))

	return *result
}
