package sqlitedb

import (
	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3"
	"github.com/pansachin/lilurl/config"
)

// Intialize datadabase connection
func Initialize(cfg *config.Config) (*sqlx.DB, error) {
	// Open a connection to the SQLite database
	db, err := sqlx.Open(cfg.DB.Instance, "./schema/lilurl.db")
	if err != nil {
		return db, err
	}
	return db, nil
}
