package store

import (
	"context"
	"log/slog"
	"testing"
	"time"

	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3"
)

const createTableSQL = `
CREATE TABLE IF NOT EXISTS urls(
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    long_url VARCHAR(255) NOT NULL,
    short VARCHAR(7) NOT NULL,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at DATETIME DEFAULT NULL
);`

func setupTestDB(t *testing.T) *sqlx.DB {
	t.Helper()
	db, err := sqlx.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("failed to open test db: %v", err)
	}
	if _, err := db.Exec(createTableSQL); err != nil {
		t.Fatalf("failed to create table: %v", err)
	}
	t.Cleanup(func() { db.Close() })
	return db
}

func newTestStore(t *testing.T) *Store {
	t.Helper()
	db := setupTestDB(t)
	logger := slog.Default()
	return New(db, logger)
}

func insertTestURL(t *testing.T, s *Store, longURL, short string) LilURL {
	t.Helper()
	now := time.Now().Truncate(time.Second)
	data := LilURL{
		Long:      longURL,
		Short:     short,
		CreatedAt: now,
		UpdatedAt: now,
	}
	q := `INSERT INTO urls (long_url, short, created_at, updated_at) VALUES (:long_url, :short, :created_at, :updated_at)`
	res, err := sqlx.NamedExecContext(context.Background(), s.db, q, &data)
	if err != nil {
		t.Fatalf("failed to insert test url: %v", err)
	}
	id, _ := res.LastInsertId()
	data.ID = int(id)
	return data
}

func TestGetByID(t *testing.T) {
	s := newTestStore(t)
	inserted := insertTestURL(t, s, "https://example.com", "abc1234")

	tests := []struct {
		name      string
		id        int64
		wantFound bool
	}{
		{"existing record", int64(inserted.ID), true},
		{"non-existent record", 999, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := s.GetByID(tt.id)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if tt.wantFound && result.ID == 0 {
				t.Error("expected record to be found, got zero ID")
			}
			if !tt.wantFound && result.ID != 0 {
				t.Errorf("expected no record, got ID %d", result.ID)
			}
			if tt.wantFound {
				if result.Long != "https://example.com" {
					t.Errorf("expected long_url %q, got %q", "https://example.com", result.Long)
				}
				if result.Short != "abc1234" {
					t.Errorf("expected short %q, got %q", "abc1234", result.Short)
				}
			}
		})
	}
}

func TestGetByShortURL(t *testing.T) {
	s := newTestStore(t)
	insertTestURL(t, s, "https://example.com", "xyz7890")

	tests := []struct {
		name      string
		short     string
		wantFound bool
	}{
		{"existing record", "xyz7890", true},
		{"non-existent record", "nope123", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := s.GetByShortURL(tt.short)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if tt.wantFound && result.ID == 0 {
				t.Error("expected record to be found, got zero ID")
			}
			if !tt.wantFound && result.ID != 0 {
				t.Errorf("expected no record, got ID %d", result.ID)
			}
		})
	}
}

func TestCreate(t *testing.T) {
	s := newTestStore(t)
	now := time.Now().Truncate(time.Second)

	data := LilURL{
		Long:      "https://example.com/create-test",
		CreatedAt: now,
		UpdatedAt: now,
	}

	result, err := s.Create(data)
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	if result.ID == 0 {
		t.Error("expected non-zero ID")
	}
	if result.Long != "https://example.com/create-test" {
		t.Errorf("expected long_url %q, got %q", "https://example.com/create-test", result.Long)
	}
	if result.Short == "" {
		t.Error("expected non-empty short URL")
	}
	if len(result.Short) != 7 {
		t.Errorf("expected short URL length 7, got %d", len(result.Short))
	}
}

func TestUpdate(t *testing.T) {
	s := newTestStore(t)
	inserted := insertTestURL(t, s, "https://example.com", "old1234")

	now := time.Now().Truncate(time.Second)
	err := s.Update(LilURL{
		ID:        inserted.ID,
		Short:     "new5678",
		UpdatedAt: now,
	})
	if err != nil {
		t.Fatalf("Update failed: %v", err)
	}

	result, err := s.GetByID(int64(inserted.ID))
	if err != nil {
		t.Fatalf("GetByID after update failed: %v", err)
	}
	if result.Short != "new5678" {
		t.Errorf("expected short %q after update, got %q", "new5678", result.Short)
	}
}

func TestDelete(t *testing.T) {
	s := newTestStore(t)

	t.Run("successful soft delete", func(t *testing.T) {
		inserted := insertTestURL(t, s, "https://example.com/del", "del1234")

		err := s.Delete(int64(inserted.ID))
		if err != nil {
			t.Fatalf("Delete failed: %v", err)
		}

		// Verify record is excluded from GetByID
		result, err := s.GetByID(int64(inserted.ID))
		if err != nil {
			t.Fatalf("GetByID after delete failed: %v", err)
		}
		if result.ID != 0 {
			t.Error("expected deleted record to not be found by GetByID")
		}

		// Verify record is excluded from GetByShortURL
		result, err = s.GetByShortURL("del1234")
		if err != nil {
			t.Fatalf("GetByShortURL after delete failed: %v", err)
		}
		if result.ID != 0 {
			t.Error("expected deleted record to not be found by GetByShortURL")
		}
	})

	t.Run("delete non-existent record", func(t *testing.T) {
		err := s.Delete(999)
		if err != ErrNotFound {
			t.Errorf("expected ErrNotFound, got %v", err)
		}
	})

	t.Run("double delete returns not found", func(t *testing.T) {
		inserted := insertTestURL(t, s, "https://example.com/double", "dbl1234")
		if err := s.Delete(int64(inserted.ID)); err != nil {
			t.Fatalf("first Delete failed: %v", err)
		}
		err := s.Delete(int64(inserted.ID))
		if err != ErrNotFound {
			t.Errorf("expected ErrNotFound on double delete, got %v", err)
		}
	})
}

func TestSoftDeleteEnforcement(t *testing.T) {
	s := newTestStore(t)

	// Insert two records, delete one
	kept := insertTestURL(t, s, "https://example.com/keep", "keep123")
	deleted := insertTestURL(t, s, "https://example.com/del", "delt123")
	if err := s.Delete(int64(deleted.ID)); err != nil {
		t.Fatalf("Delete failed: %v", err)
	}

	// GetByID should return the kept record
	result, err := s.GetByID(int64(kept.ID))
	if err != nil {
		t.Fatalf("GetByID for kept record failed: %v", err)
	}
	if result.ID != kept.ID {
		t.Errorf("expected kept record ID %d, got %d", kept.ID, result.ID)
	}

	// GetByID should not return the deleted record
	result, err = s.GetByID(int64(deleted.ID))
	if err != nil {
		t.Fatalf("GetByID for deleted record failed: %v", err)
	}
	if result.ID != 0 {
		t.Error("expected deleted record to not be returned")
	}

	// GetByShortURL should not return the deleted record
	result, err = s.GetByShortURL("delt123")
	if err != nil {
		t.Fatalf("GetByShortURL for deleted record failed: %v", err)
	}
	if result.ID != 0 {
		t.Error("expected deleted record to not be returned by short URL")
	}
}
