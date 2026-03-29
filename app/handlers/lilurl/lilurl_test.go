package lilurl

import (
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/limiter"
	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3"
	"github.com/pansachin/lilurl/config"
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

func setupTestApp(t *testing.T) *fiber.App {
	t.Helper()
	db, err := sqlx.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("failed to open test db: %v", err)
	}
	if _, err := db.Exec(createTableSQL); err != nil {
		t.Fatalf("failed to create table: %v", err)
	}
	t.Cleanup(func() { db.Close() })

	app := fiber.New()
	logger := slog.Default()
	h := NewHandler(db, logger)
	rl := &config.RateLimit{
		Max:              60,
		WindowSecs:       60,
		CreateMax:        10,
		CreateWindowSecs: 60,
	}

	app.Get("/health", func(c fiber.Ctx) error {
		return c.Status(fiber.StatusOK).JSON(fiber.Map{"status": "ok"})
	})
	app.Get("/:lilurl", h.Get)
	app.Get("/api/v1/short/:lilurl", h.GetByShortURL)
	app.Get("/api/v1/id/:id", h.GetByID)
	createLimiter := limiter.New(limiter.Config{
		Max:        rl.CreateMax,
		Expiration: time.Duration(rl.CreateWindowSecs) * time.Second,
		KeyGenerator: func(c fiber.Ctx) string {
			return "create:" + c.IP()
		},
		LimiterMiddleware: limiter.SlidingWindow{},
	})
	app.Post("/api/v1/lilurl", createLimiter, h.Create)
	app.Delete("/api/v1/:id", h.Delete)

	return app
}

func parseJSON(t *testing.T, body io.Reader) map[string]any {
	t.Helper()
	var result map[string]any
	if err := json.NewDecoder(body).Decode(&result); err != nil {
		t.Fatalf("failed to parse JSON: %v", err)
	}
	return result
}

func createTestURL(t *testing.T, app *fiber.App, longURL string) map[string]any {
	t.Helper()
	req, _ := http.NewRequest("POST", "/api/v1/lilurl", strings.NewReader(`{"long_url": "`+longURL+`"}`))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("failed to create test URL: %v", err)
	}
	if resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		t.Fatalf("expected 201 on create, got %d: %s", resp.StatusCode, string(body))
	}
	data := parseJSON(t, resp.Body)
	result, ok := data["result"].(map[string]any)
	if !ok {
		t.Fatal("expected result object in create response")
	}
	return result
}

func TestHealthCheck(t *testing.T) {
	app := setupTestApp(t)

	req, _ := http.NewRequest("GET", "/health", nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected status 200, got %d", resp.StatusCode)
	}

	data := parseJSON(t, resp.Body)
	if data["status"] != "ok" {
		t.Errorf("expected status 'ok', got %v", data["status"])
	}
}

func TestCreate(t *testing.T) {
	app := setupTestApp(t)

	t.Run("valid URL", func(t *testing.T) {
		req, _ := http.NewRequest("POST", "/api/v1/lilurl", strings.NewReader(`{"long_url": "https://example.com/valid"}`))
		req.Header.Set("Content-Type", "application/json")
		resp, err := app.Test(req)
		if err != nil {
			t.Fatalf("request failed: %v", err)
		}

		if resp.StatusCode != http.StatusCreated {
			body, _ := io.ReadAll(resp.Body)
			t.Fatalf("expected status 201, got %d: %s", resp.StatusCode, string(body))
		}

		data := parseJSON(t, resp.Body)
		result, ok := data["result"].(map[string]any)
		if !ok {
			t.Fatal("expected result object")
		}
		if result["long_url"] != "https://example.com/valid" {
			t.Errorf("expected long_url 'https://example.com/valid', got %v", result["long_url"])
		}
		if result["short"] == nil || result["short"] == "" {
			t.Error("expected non-empty short URL")
		}
	})

	t.Run("invalid URL", func(t *testing.T) {
		req, _ := http.NewRequest("POST", "/api/v1/lilurl", strings.NewReader(`{"long_url": "not-a-url"}`))
		req.Header.Set("Content-Type", "application/json")
		resp, err := app.Test(req)
		if err != nil {
			t.Fatalf("request failed: %v", err)
		}

		if resp.StatusCode != http.StatusInternalServerError {
			t.Errorf("expected status 500 for invalid URL, got %d", resp.StatusCode)
		}
	})

	t.Run("missing body", func(t *testing.T) {
		req, _ := http.NewRequest("POST", "/api/v1/lilurl", strings.NewReader(`{}`))
		req.Header.Set("Content-Type", "application/json")
		resp, err := app.Test(req)
		if err != nil {
			t.Fatalf("request failed: %v", err)
		}

		if resp.StatusCode == http.StatusCreated {
			t.Error("expected error status for empty body, got 201")
		}
	})
}

func TestGetByID(t *testing.T) {
	app := setupTestApp(t)
	created := createTestURL(t, app, "https://example.com/getbyid")
	id := created["id"].(float64)

	t.Run("existing record", func(t *testing.T) {
		req, _ := http.NewRequest("GET", fmt.Sprintf("/api/v1/id/%d", int(id)), nil)
		resp, err := app.Test(req)
		if err != nil {
			t.Fatalf("request failed: %v", err)
		}

		if resp.StatusCode != http.StatusOK {
			body, _ := io.ReadAll(resp.Body)
			t.Fatalf("expected status 200, got %d: %s", resp.StatusCode, string(body))
		}

		data := parseJSON(t, resp.Body)
		result := data["result"].(map[string]any)
		if result["long_url"] != "https://example.com/getbyid" {
			t.Errorf("expected long_url, got %v", result["long_url"])
		}
	})

	t.Run("non-existent ID", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/api/v1/id/999", nil)
		resp, err := app.Test(req)
		if err != nil {
			t.Fatalf("request failed: %v", err)
		}
		if resp.StatusCode != http.StatusNotFound {
			t.Errorf("expected 404, got %d", resp.StatusCode)
		}
	})

	t.Run("non-numeric ID", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/api/v1/id/abc", nil)
		resp, err := app.Test(req)
		if err != nil {
			t.Fatalf("request failed: %v", err)
		}
		if resp.StatusCode != http.StatusBadRequest {
			t.Errorf("expected 400, got %d", resp.StatusCode)
		}
	})
}

func TestGetByShortURL(t *testing.T) {
	app := setupTestApp(t)
	created := createTestURL(t, app, "https://example.com/getbyshort")
	short := created["short"].(string)

	t.Run("existing record", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/api/v1/short/"+short, nil)
		resp, err := app.Test(req)
		if err != nil {
			t.Fatalf("request failed: %v", err)
		}
		if resp.StatusCode != http.StatusOK {
			body, _ := io.ReadAll(resp.Body)
			t.Fatalf("expected 200, got %d: %s", resp.StatusCode, string(body))
		}

		data := parseJSON(t, resp.Body)
		result := data["result"].(map[string]any)
		if result["long_url"] != "https://example.com/getbyshort" {
			t.Errorf("expected long_url, got %v", result["long_url"])
		}
	})

	t.Run("non-existent short URL", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/api/v1/short/nope123", nil)
		resp, err := app.Test(req)
		if err != nil {
			t.Fatalf("request failed: %v", err)
		}
		if resp.StatusCode != http.StatusNotFound {
			t.Errorf("expected 404, got %d", resp.StatusCode)
		}
	})
}

func TestDelete(t *testing.T) {
	app := setupTestApp(t)

	t.Run("successful delete", func(t *testing.T) {
		created := createTestURL(t, app, "https://example.com/delete-test")
		id := int(created["id"].(float64))

		req, _ := http.NewRequest("DELETE", fmt.Sprintf("/api/v1/%d", id), nil)
		resp, err := app.Test(req)
		if err != nil {
			t.Fatalf("request failed: %v", err)
		}
		if resp.StatusCode != http.StatusNoContent {
			body, _ := io.ReadAll(resp.Body)
			t.Errorf("expected 204, got %d: %s", resp.StatusCode, string(body))
		}
	})

	t.Run("delete non-existent", func(t *testing.T) {
		req, _ := http.NewRequest("DELETE", "/api/v1/999", nil)
		resp, err := app.Test(req)
		if err != nil {
			t.Fatalf("request failed: %v", err)
		}
		if resp.StatusCode != http.StatusNotFound {
			t.Errorf("expected 404, got %d", resp.StatusCode)
		}
	})
}

func TestGetRedirect(t *testing.T) {
	app := setupTestApp(t)
	created := createTestURL(t, app, "https://example.com/redirect-test")
	short := created["short"].(string)

	t.Run("valid redirect", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/"+short, nil)
		resp, err := app.Test(req)
		if err != nil {
			t.Fatalf("request failed: %v", err)
		}
		if resp.StatusCode != http.StatusTemporaryRedirect {
			body, _ := io.ReadAll(resp.Body)
			t.Fatalf("expected 307, got %d: %s", resp.StatusCode, string(body))
		}
		location := resp.Header.Get("Location")
		if location != "https://example.com/redirect-test" {
			t.Errorf("expected Location header 'https://example.com/redirect-test', got %q", location)
		}
	})

	t.Run("non-existent short URL", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/nope123", nil)
		resp, err := app.Test(req)
		if err != nil {
			t.Fatalf("request failed: %v", err)
		}
		if resp.StatusCode != http.StatusNotFound {
			t.Errorf("expected 404, got %d", resp.StatusCode)
		}
	})
}

func TestDeleteThenGet(t *testing.T) {
	app := setupTestApp(t)
	created := createTestURL(t, app, "https://example.com/soft-delete")
	short := created["short"].(string)
	id := int(created["id"].(float64))

	// Delete the record
	delReq, _ := http.NewRequest("DELETE", fmt.Sprintf("/api/v1/%d", id), nil)
	resp, err := app.Test(delReq)
	if err != nil {
		t.Fatalf("delete request failed: %v", err)
	}
	if resp.StatusCode != http.StatusNoContent {
		t.Fatalf("expected 204 on delete, got %d", resp.StatusCode)
	}

	// GET by short URL should return 404
	req, _ := http.NewRequest("GET", "/api/v1/short/"+short, nil)
	resp, err = app.Test(req)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	if resp.StatusCode != http.StatusNotFound {
		t.Errorf("expected 404 after delete, got %d", resp.StatusCode)
	}

	// Redirect should return 404
	req, _ = http.NewRequest("GET", "/"+short, nil)
	resp, err = app.Test(req)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	if resp.StatusCode != http.StatusNotFound {
		t.Errorf("expected 404 redirect after delete, got %d", resp.StatusCode)
	}
}
