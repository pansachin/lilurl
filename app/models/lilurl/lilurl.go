package model

import (
	"unsafe"

	validate "github.com/go-playground/validator/v10"
	"github.com/jmoiron/sqlx"
	store "github.com/pansachin/lilurl/app/models/lilurl/db"
)

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

// Create creates a new lilurl in the database
func (c *Core) Update(payload UpdateLilURL) error {
	if err := validate.New().Struct(payload); err != nil {
		return err
	}

	data := store.LilURL{
		ID:        payload.ID,
		Short:     payload.Short,
		UpdatedAt: payload.UpdatedAt,
	}

	err := c.db.Update(data)

	return err
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
