package model

import "time"

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
	Short     string    `json:"short"`
	CretedAt  time.Time `json:"created_at" validate:"required"`
	UpdatedAt time.Time `json:"updated_at" validate:"required"`
}

// UpdateLilURL model
type UpdateLilURL struct {
	ID        int       `json:"id" validate:"required"`
	Short     string    `json:"short" validate:"required"`
	UpdatedAt time.Time `json:"updated_at" validate:"required"`
}
