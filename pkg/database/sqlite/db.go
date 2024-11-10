package sqlitedb

import (
	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3"
)

func Initialize() (*sqlx.DB, error) {
	// Open a connection to the SQLite database
	db, err := sqlx.Open("sqlite3", "./schema/lilurl.db")
	if err != nil {
		return db, err
	}
	return db, nil
}
