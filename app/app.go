package app

import (
	"database/sql"
)

// App config
type Config struct {
	// Database instance
	DB *sql.DB
}
