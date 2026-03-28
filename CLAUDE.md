# CLAUDE.md

## Project Overview

LilURL is a URL shortener service built with Go and the Fiber v3 web framework, using SQLite as the database backend.

## Build & Run

```bash
# Install dependencies
go mod download

# Run database migrations (requires dbmate)
make migrate

# Build binary (CGO required for SQLite)
CGO_ENABLED=1 go build -o lilurl .

# Run
./lilurl

# Docker
make build   # build image
make run     # run container
docker-compose up --build  # full stack with nginx proxy
```

App runs on `http://localhost:3000` by default.

## Test

```bash
make test          # or: go test ./...
go test -race ./...   # with race detector
go test -cover ./...  # with coverage
```

Test files:
- `config/config_test.go` — config tests
- `internal/pkg/generator/generator_test.go` — generator tests and benchmarks

## Lint / Format

No linter is configured. Use standard Go tooling:

```bash
go fmt ./...
go vet ./...
```

## Project Structure

```
main.go                          — entry point
config/                          — config loading (cleanenv + config.yaml)
app/handlers/                    — HTTP handlers and route registration
app/models/lilurl/               — data models and business logic
app/models/lilurl/db/            — SQLite CRUD operations (sqlx)
internal/pkg/generator/          — URL shortening algorithms (base62, sha256)
pkg/database/sqlite/             — SQLite initialization
pkg/log/                         — structured logging (slog)
schema/migrations/               — dbmate migrations
schema/schema.sql                — current DB schema
```

## Key Technical Details

- **Go 1.23+** with CGO enabled (required by `mattn/go-sqlite3`)
- **Fiber v3** (beta) web framework on FastHTTP
- **SQLite** database stored at `schema/lilurl.db`
- Config loaded from `config.yaml` via `cleanenv`; DB config in `.env`
- CORS allows `http://localhost:5173` (Vite dev server)
- Validation via `go-playground/validator/v10`

## API Endpoints

- `POST /api/v1/lilurl` — create short URL (JSON body with `long_url`)
- `GET /:lilurl` — redirect to original URL (301)
- `GET /api/v1/:lilurl` — get URL details by short code
- `GET /api/v1/:id` — get URL details by ID

## CI

GitHub Actions (`.github/workflows/ci.yml`) runs tests and builds/pushes Docker images to Google Artifact Registry.
