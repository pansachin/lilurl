package lilurl_test

import (
	"bytes"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	"github.com/gofiber/fiber/v3"
	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3"
	handler "github.com/pansachin/lilurl/app/handlers/lilurl"
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

func newTestApp(t *testing.T) *fiber.App {
	t.Helper()
	db := newTestDB(t)
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	h := handler.NewHandler(db, logger)

	app := fiber.New()
	app.Get("/health", func(c fiber.Ctx) error {
		return c.Status(fiber.StatusOK).JSON(fiber.Map{"status": "ok"})
	})
	app.Post("/api/v1/lilurl", h.Create)
	app.Get("/api/v1/short/:lilurl", h.GetByShortURL)
	app.Get("/api/v1/id/:id", h.GetByID)
	app.Delete("/api/v1/:id", h.Delete)
	app.Get("/:lilurl", h.Get)
	return app
}

// createShortURL is a helper that POSTs a URL and returns the result map.
func createShortURL(t *testing.T, app *fiber.App, longURL string) map[string]interface{} {
	t.Helper()
	body, _ := json.Marshal(map[string]string{"long_url": longURL})
	req := httptest.NewRequest(http.MethodPost, "/api/v1/lilurl", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("app.Test() error = %v", err)
	}
	if resp.StatusCode != http.StatusCreated {
		b, _ := io.ReadAll(resp.Body)
		t.Fatalf("createShortURL() status = %d, want 201, body = %s", resp.StatusCode, b)
	}
	var envelope struct {
		Result map[string]interface{} `json:"result"`
	}
	json.NewDecoder(resp.Body).Decode(&envelope)
	return envelope.Result
}

func TestHealth(t *testing.T) {
	app := newTestApp(t)
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("app.Test() error = %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Errorf("GET /health status = %d, want 200", resp.StatusCode)
	}
}

func TestCreate_ValidInput(t *testing.T) {
	app := newTestApp(t)
	body := bytes.NewBufferString(`{"long_url":"https://example.com"}`)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/lilurl", body)
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("app.Test() error = %v", err)
	}
	if resp.StatusCode != http.StatusCreated {
		t.Errorf("Create() status = %d, want 201", resp.StatusCode)
	}

	var envelope struct {
		Result map[string]interface{} `json:"result"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&envelope); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if envelope.Result["short"] == "" {
		t.Error("Create() short is empty in response")
	}
}

func TestCreate_MalformedJSON(t *testing.T) {
	app := newTestApp(t)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/lilurl", bytes.NewBufferString(`{invalid`))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("app.Test() error = %v", err)
	}
	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("Create() malformed JSON status = %d, want 400", resp.StatusCode)
	}
}

func TestCreate_InvalidURL(t *testing.T) {
	app := newTestApp(t)
	body := bytes.NewBufferString(`{"long_url":"not-a-valid-url"}`)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/lilurl", body)
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("app.Test() error = %v", err)
	}
	if resp.StatusCode != http.StatusInternalServerError {
		t.Errorf("Create() invalid URL status = %d, want 500", resp.StatusCode)
	}
}

func TestGetByShortURL_Found(t *testing.T) {
	app := newTestApp(t)
	created := createShortURL(t, app, "https://example.com")
	short := created["short"].(string)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/short/"+short, nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("app.Test() error = %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Errorf("GetByShortURL() status = %d, want 200", resp.StatusCode)
	}
}

func TestGetByShortURL_NotFound(t *testing.T) {
	app := newTestApp(t)
	req := httptest.NewRequest(http.MethodGet, "/api/v1/short/doesntexist", nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("app.Test() error = %v", err)
	}
	if resp.StatusCode != http.StatusNotFound {
		t.Errorf("GetByShortURL() not-found status = %d, want 404", resp.StatusCode)
	}
}

func TestGetByID_Found(t *testing.T) {
	app := newTestApp(t)
	created := createShortURL(t, app, "https://example.com")
	id := strconv.Itoa(int(created["id"].(float64)))

	req := httptest.NewRequest(http.MethodGet, "/api/v1/id/"+id, nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("app.Test() error = %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Errorf("GetByID() status = %d, want 200", resp.StatusCode)
	}
}

func TestGetByID_NotFound(t *testing.T) {
	app := newTestApp(t)
	req := httptest.NewRequest(http.MethodGet, "/api/v1/id/99999", nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("app.Test() error = %v", err)
	}
	if resp.StatusCode != http.StatusNotFound {
		t.Errorf("GetByID() not-found status = %d, want 404", resp.StatusCode)
	}
}

func TestGetByID_InvalidID(t *testing.T) {
	app := newTestApp(t)
	req := httptest.NewRequest(http.MethodGet, "/api/v1/id/notanumber", nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("app.Test() error = %v", err)
	}
	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("GetByID() invalid ID status = %d, want 400", resp.StatusCode)
	}
}

func TestGet_Redirect(t *testing.T) {
	app := newTestApp(t)
	created := createShortURL(t, app, "https://example.com")
	short := created["short"].(string)

	req := httptest.NewRequest(http.MethodGet, "/"+short, nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("app.Test() error = %v", err)
	}
	if resp.StatusCode != http.StatusTemporaryRedirect {
		t.Errorf("Get() status = %d, want 307", resp.StatusCode)
	}
	if loc := resp.Header.Get("Location"); loc != "https://example.com" {
		t.Errorf("Get() Location = %q, want %q", loc, "https://example.com")
	}
}

func TestGet_NotFound(t *testing.T) {
	app := newTestApp(t)
	req := httptest.NewRequest(http.MethodGet, "/doesnotexist", nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("app.Test() error = %v", err)
	}
	if resp.StatusCode != http.StatusNotFound {
		t.Errorf("Get() not-found status = %d, want 404", resp.StatusCode)
	}
}

func TestDelete_Success(t *testing.T) {
	app := newTestApp(t)
	created := createShortURL(t, app, "https://example.com")
	id := strconv.Itoa(int(created["id"].(float64)))

	req := httptest.NewRequest(http.MethodDelete, "/api/v1/"+id, nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("app.Test() error = %v", err)
	}
	if resp.StatusCode != http.StatusNoContent {
		t.Errorf("Delete() status = %d, want 204", resp.StatusCode)
	}
}

func TestDelete_NotFound(t *testing.T) {
	app := newTestApp(t)
	req := httptest.NewRequest(http.MethodDelete, "/api/v1/99999", nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("app.Test() error = %v", err)
	}
	if resp.StatusCode != http.StatusNotFound {
		t.Errorf("Delete() not-found status = %d, want 404", resp.StatusCode)
	}
}

func TestDelete_InvalidID(t *testing.T) {
	app := newTestApp(t)
	req := httptest.NewRequest(http.MethodDelete, "/api/v1/notanumber", nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("app.Test() error = %v", err)
	}
	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("Delete() invalid ID status = %d, want 400", resp.StatusCode)
	}
}
