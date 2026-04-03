package store_test

import (
	"errors"
	"io"
	"log/slog"
	"testing"
	"time"

	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3"
	store "github.com/pansachin/lilurl/app/models/lilurl/db"
)

const testSchema = `CREATE TABLE urls (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	long_url VARCHAR(255) NOT NULL,
	short VARCHAR(7) NOT NULL,
	created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
	updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
	deleted_at DATETIME DEFAULT NULL
)`

func newTestDB(t *testing.T) *sqlx.DB {
	t.Helper()
	db, err := sqlx.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	db.SetMaxOpenConns(1)
	if _, err := db.Exec(testSchema); err != nil {
		t.Fatalf("create schema: %v", err)
	}
	t.Cleanup(func() { db.Close() })
	return db
}

func newTestStore(t *testing.T) *store.Store {
	t.Helper()
	return store.New(newTestDB(t), slog.New(slog.NewTextHandler(io.Discard, nil)))
}

func newEntry() store.LilURL {
	now := time.Now().Truncate(time.Second)
	return store.LilURL{
		Long:      "https://example.com",
		CreatedAt: now,
		UpdatedAt: now,
	}
}

func TestStore_Create(t *testing.T) {
	s := newTestStore(t)
	result, err := s.Create(newEntry())
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}
	if result.ID == 0 {
		t.Error("Create() ID = 0, want non-zero")
	}
	if result.Long != "https://example.com" {
		t.Errorf("Create() Long = %q, want %q", result.Long, "https://example.com")
	}
	if result.Short == "" {
		t.Error("Create() Short is empty")
	}
}

func TestStore_GetByID(t *testing.T) {
	s := newTestStore(t)
	created, err := s.Create(newEntry())
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	got, err := s.GetByID(int64(created.ID))
	if err != nil {
		t.Fatalf("GetByID() error = %v", err)
	}
	if got.ID != created.ID {
		t.Errorf("GetByID() ID = %d, want %d", got.ID, created.ID)
	}
	if got.Short != created.Short {
		t.Errorf("GetByID() Short = %q, want %q", got.Short, created.Short)
	}
	if got.Long != created.Long {
		t.Errorf("GetByID() Long = %q, want %q", got.Long, created.Long)
	}
}

func TestStore_GetByID_NotFound(t *testing.T) {
	s := newTestStore(t)
	got, err := s.GetByID(99999)
	if err != nil {
		t.Fatalf("GetByID() error = %v", err)
	}
	if got.ID != 0 {
		t.Errorf("GetByID() not-found should return zero ID, got %d", got.ID)
	}
}

func TestStore_GetByShortURL(t *testing.T) {
	s := newTestStore(t)
	created, err := s.Create(newEntry())
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	got, err := s.GetByShortURL(created.Short)
	if err != nil {
		t.Fatalf("GetByShortURL() error = %v", err)
	}
	if got.ID != created.ID {
		t.Errorf("GetByShortURL() ID = %d, want %d", got.ID, created.ID)
	}
	if got.Long != created.Long {
		t.Errorf("GetByShortURL() Long = %q, want %q", got.Long, created.Long)
	}
}

func TestStore_GetByShortURL_NotFound(t *testing.T) {
	s := newTestStore(t)
	got, err := s.GetByShortURL("notexist")
	if err != nil {
		t.Fatalf("GetByShortURL() error = %v", err)
	}
	if got.ID != 0 {
		t.Errorf("GetByShortURL() not-found should return zero ID, got %d", got.ID)
	}
}

func TestStore_Delete(t *testing.T) {
	s := newTestStore(t)
	created, err := s.Create(newEntry())
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	if err := s.Delete(int64(created.ID)); err != nil {
		t.Fatalf("Delete() error = %v", err)
	}

	got, err := s.GetByID(int64(created.ID))
	if err != nil {
		t.Fatalf("GetByID() after Delete() error = %v", err)
	}
	if got.ID != 0 {
		t.Error("GetByID() returned deleted record")
	}
}

func TestStore_Delete_NotFound(t *testing.T) {
	s := newTestStore(t)
	err := s.Delete(99999)
	if !errors.Is(err, store.ErrNotFound) {
		t.Errorf("Delete() error = %v, want ErrNotFound", err)
	}
}

func TestStore_Delete_AlreadyDeleted(t *testing.T) {
	s := newTestStore(t)
	created, err := s.Create(newEntry())
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	if err := s.Delete(int64(created.ID)); err != nil {
		t.Fatalf("first Delete() error = %v", err)
	}

	err = s.Delete(int64(created.ID))
	if !errors.Is(err, store.ErrNotFound) {
		t.Errorf("second Delete() error = %v, want ErrNotFound", err)
	}
}

func TestStore_GetByID_ExcludesDeleted(t *testing.T) {
	s := newTestStore(t)
	created, err := s.Create(newEntry())
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	if err := s.Delete(int64(created.ID)); err != nil {
		t.Fatalf("Delete() error = %v", err)
	}

	got, err := s.GetByID(int64(created.ID))
	if err != nil {
		t.Fatalf("GetByID() error = %v", err)
	}
	if got.ID != 0 {
		t.Error("GetByID() returned deleted record")
	}
}

func TestStore_GetByShortURL_ExcludesDeleted(t *testing.T) {
	s := newTestStore(t)
	created, err := s.Create(newEntry())
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	if err := s.Delete(int64(created.ID)); err != nil {
		t.Fatalf("Delete() error = %v", err)
	}

	got, err := s.GetByShortURL(created.Short)
	if err != nil {
		t.Fatalf("GetByShortURL() error = %v", err)
	}
	if got.ID != 0 {
		t.Error("GetByShortURL() returned deleted record")
	}
}
