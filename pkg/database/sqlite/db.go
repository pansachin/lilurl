package sqlitedb

import (
	"database/sql"
	"fmt"

	_ "github.com/mattn/go-sqlite3"
)

func Initialize() (*sql.DB, error) {
	// Open a connection to the SQLite database
	db, err := sql.Open("sqlite3", "../../../data/example.db")
	if err != nil {
		return db, err
	}
	return db, nil
}

func CreateTable(db *sql.DB) error {
	// Create a new table
	createTableSQL := `CREATE TABLE IF NOT EXISTS users (
		"id" integer NOT NULL PRIMARY KEY AUTOINCREMENT,		
		"name" TEXT,
		"email" TEXT);`
	_, err := db.Exec(createTableSQL)
	if err != nil {
		return err
	}
	fmt.Println("Table created successfully")
	return nil
}

func Insert(db *sql.DB, param ...string) error {
	// Insert some data
	insertUserSQL := `INSERT INTO users (name, email) VALUES (?, ?)`
	_, err := db.Exec(insertUserSQL, param[0], param[1])
	if err != nil {
		return err
	}
	fmt.Println("User inserted successfully")

	return nil
}

func Query(db *sql.DB) error {
	// Query the data
	rows, err := db.Query("SELECT id, name, email FROM users")
	if err != nil {
		return err
	}
	defer rows.Close()

	// Iterate over the result set
	for rows.Next() {
		var id int
		var name, email string
		err = rows.Scan(&id, &name, &email)
		if err != nil {
			return err
		}
		fmt.Printf("ID: %d, Name: %s, Email: %s\n", id, name, email)
	}

	// Handle any errors from iterating over rows
	err = rows.Err()
	if err != nil {
		return err
	}

	return nil
}
