package store

import (
	"context"
	"log/slog"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/pansachin/lilurl/internal/pkg/generator"
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
	db        sqlx.ExtContext
	logger    *slog.Logger
	generator *generator.ShortURLGenerator
}

// New creates a new model
func New(db *sqlx.DB, logger *slog.Logger) *Store {
	return &Store{
		db:        db,
		logger:    logger,
		generator: generator.NewShortURLGenerator(7, 10000), // 7 chars, cache up to 10k entries
	}
}

// Create creates a new lilurl in the database
func (s *Store) Create(data LilURL) (LilURL, error) {
	q := `
	INSERT INTO
		urls (long_url, short, created_at, updated_at, deleted_at)
	VALUES
		(:long_url, :short, :created_at, :updated_at, :deleted_at)`

	// Log query
	str, values, err := sqlx.Named(q, &data)
	s.logger.Debug("insert query", "str", str, "values", values, "err", err)

	res, err := sqlx.NamedExecContext(context.Background(), s.db, q, &data)
	if err != nil {
		return LilURL{}, err
	}

	id, err := res.LastInsertId()
	if err != nil {
		return LilURL{}, err
	}

	// Use the generator to create a unique short URL
	checkFunc := func(short string) (bool, error) {
		result, err := s.GetByShortURL(short)
		if err != nil {
			return false, err
		}
		return result.ID != 0, nil
	}

	shortURL, err := s.generator.Generate(data.Long, checkFunc)
	if err != nil {
		return LilURL{}, err
	}
	data.Short = shortURL
	s.logger.Debug("generated short url", "short", shortURL)

	// Update record with short url ID
	data.ID = int(id)
	if err := s.Update(data); err != nil {
		return LilURL{}, err
	}

	return s.GetByID(id)
}

// Update updates an existing lilurl in the database
func (s *Store) Update(data LilURL) error {
	q := `
		UPDATE
			urls
		SET
			short = :short,
			updated_at = :updated_at
		WHERE
			id = :id`

	// Log query
	str, values, err := sqlx.Named(q, &data)
	s.logger.Debug("updare data by id", "str", str, "values", values, "err", err)

	_, err = sqlx.NamedExecContext(context.Background(), s.db, q, &data)

	return err
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

	// Log query
	str, values, err := sqlx.Named(q, &args)
	s.logger.Debug("get data by id", "str", str, "values", values, "err", err)

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

	// Log query
	str, values, err := sqlx.Named(q, &args)
	s.logger.Debug("get data by short url", "str", str, "values", values, "err", err)

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
