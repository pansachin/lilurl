package store

import (
	"context"
	"time"

	"github.com/jmoiron/sqlx"
)

// LilURL model
type LilURL struct {
	ID        int        `db:"id"`
	Long      string     `db:"long_url"`
	Short     string     `db:"short"`
	CreatedAt time.Time  `db:"created_at"`
	UpdatedAt time.Time  `db:"updated_at"`
	DeleredAt *time.Time `db:"deleted_at"`
}

// Store
type Store struct {
	db sqlx.ExtContext
}

// New creates a new model
func New(db *sqlx.DB) *Store {
	return &Store{
		db: db,
	}
}

// Create creates a new lilurl in the database
func (s *Store) Create(data LilURL) (LilURL, error) {
	q := `
	INSERT INTO
		urls (long_url, short, created_at, updated_at, deleted_at)
	VALUES
		(:long_url, :short, :created_at, :updated_at, :deleted_at)`

	res, err := sqlx.NamedExecContext(context.Background(), s.db, q, &data)
	if err != nil {
		return LilURL{}, err
	}

	id, err := res.LastInsertId()
	if err != nil {
		return LilURL{}, err
	}

	return s.GetByID(id)
}

// GetByID retrieves a lilurl by its ID
func (s *Store) GetByID(id int64) (LilURL, error) {
	var result LilURL

	args := struct {
		ID int64 `db:"id"`
	}{
		ID: id,
	}

	q := `
	SELECT
		id,
		long_url,
		short,
		created_at,
		updated_at,
		deleted_at
	FROM
		urls
	WHERE
		id = :id`

	rows, err := sqlx.NamedQueryContext(context.Background(), s.db, q, &args)
	if err != nil {
		return LilURL{}, err
	}
	defer rows.Close()
	for rows.Next() {
		if err := rows.StructScan(&result); err != nil {
			return LilURL{}, err
		}
	}

	return result, nil
}

// GetByShortURL retrieves a lilurl by its short URL
func (s *Store) GetByShortURL(short string) (LilURL, error) {
	var result LilURL

	args := struct {
		Short string `db:"short"`
	}{
		Short: short,
	}

	q := `
	SELECT
		id,
		long_url,
		short,
		created_at,
		updated_at,
		deleted_at
	FROM
		urls
	WHERE
		short = :short`

	rows, err := sqlx.NamedQueryContext(context.Background(), s.db, q, &args)
	if err != nil {
		return LilURL{}, err
	}
	defer rows.Close()
	for rows.Next() {
		if err := rows.StructScan(&result); err != nil {
			return LilURL{}, err
		}
	}

	return result, nil
}
