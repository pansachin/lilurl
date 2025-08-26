# LilURL

<p align="center">
  <img src="https://img.shields.io/badge/Go-00ADD8?style=for-the-badge&logo=go&logoColor=white" alt="Go">
  <img src="https://img.shields.io/badge/Fiber-00ACD7?style=for-the-badge&logo=fiber&logoColor=white" alt="Fiber">
  <img src="https://img.shields.io/badge/SQLite-003B57?style=for-the-badge&logo=sqlite&logoColor=white" alt="SQLite">
  <img src="https://img.shields.io/badge/Docker-2496ED?style=for-the-badge&logo=docker&logoColor=white" alt="Docker">
</p>

**LilURL** is a high-performance URL shortener service built with Go and the Fiber web framework. It provides a simple and efficient way to create shortened URLs with a clean RESTful API.

## 🚀 Features

- **Fast & Lightweight**: Built with Fiber v3 (beta) for exceptional performance
- **RESTful API**: Clean and intuitive API design
- **SQLite Database**: Lightweight, file-based database perfect for URL storage
- **Docker Support**: Easy deployment with Docker and docker-compose
- **URL Validation**: Built-in validation for URLs
- **Soft Deletion**: URLs are soft-deleted, maintaining data integrity
- **Configurable**: YAML-based configuration with environment variable support
- **Structured Logging**: Production-ready logging with slog
- **Two Shortening Algorithms**: Base62 encoding and SHA256-based generation

## 📋 Table of Contents

- [Prerequisites](#prerequisites)
- [Installation](#installation)
- [Configuration](#configuration)
- [API Documentation](#api-documentation)
- [Development](#development)
- [Testing](#testing)
- [Deployment](#deployment)
- [Project Structure](#project-structure)
- [Contributing](#contributing)
- [License](#license)

## 📦 Prerequisites

- Go 1.23.0 or higher
- Docker and Docker Compose (for containerized deployment)
- Make (for using Makefile commands)
- dbmate (for database migrations)
- CGO enabled (required for SQLite)

## 🛠️ Installation

### Clone the Repository

```bash
git clone https://github.com/pansachin/lilurl.git
cd lilurl
```

### Install Dependencies

```bash
go mod download
```

### Set Up Database

```bash
# Create .env file for database configuration
echo "DATABASE_URL=sqlite:schema/lilurl.db" > .env

# Run migrations
make migrate
```

### Build and Run Locally

```bash
# Build the binary (CGO required for SQLite)
CGO_ENABLED=1 go build -o lilurl .

# Run the application
./lilurl
```

The application will start on `http://localhost:3000` by default.

## ⚙️ Configuration

LilURL uses a layered configuration approach:

1. **Default Configuration** (`config.yaml`)
2. **Production Configuration** (`/config/config.yaml` in container)
3. **Environment Variables** (override any setting)

### Configuration File Example

```yaml
app:
  name: "lilurl"
  host: "localhost"
  port: 3000
db:
  instance: "sqlite3"
  port: "3001"
  user: "user"
  password: "password"
  database: "lilurl"
log:
  debug: true
  json: true
  colour: true
  print_routes: false
```

### Environment Variables

- `DATABASE_URL`: SQLite database connection string (e.g., `sqlite:schema/lilurl.db`)
- `APP_PORT`: Override the default port (default: 3000)
- `LOG_DEBUG`: Enable debug logging (true/false)

## 📚 API Documentation

### Endpoints

#### 1. Create Short URL

**POST** `/api/v1/lilurl`

```bash
curl -X POST http://localhost:3000/api/v1/lilurl \
  -H "Content-Type: application/json" \
  -d '{"long_url": "https://github.com/pansachin/lilurl"}'
```

**Request Body:**
```json
{
  "long_url": "https://github.com/pansachin/lilurl"
}
```

**Response:**
```json
{
  "id": 1,
  "long_url": "https://github.com/pansachin/lilurl",
  "short": "abc123d",
  "created_at": "2024-10-26T10:30:00Z"
}
```

#### 2. Redirect to Original URL

**GET** `/:lilurl`

```bash
curl -L http://localhost:3000/abc123d
```

Redirects to the original URL with a 301 status code.

#### 3. Get URL Details by Short Code

**GET** `/api/v1/:lilurl`

```bash
curl http://localhost:3000/api/v1/abc123d
```

**Response:**
```json
{
  "id": 1,
  "long_url": "https://github.com/pansachin/lilurl",
  "short": "abc123d",
  "created_at": "2024-10-26T10:30:00Z",
  "updated_at": "2024-10-26T10:30:00Z"
}
```

#### 4. Get URL Details by ID

**GET** `/api/v1/:id`

```bash
curl http://localhost:3000/api/v1/1
```

## 🔧 Development

### Quick Start with Make

```bash
# Run tests
make test

# Build Docker image
make build

# Run in Docker environment (builds and runs)
make run

# Clean up Docker containers
make clean

# Remove specific container
make rm
```

### Manual Development Commands

```bash
# Run tests with verbose output
go test -v ./...

# Run with hot reload (using air)
air

# Build for production
CGO_ENABLED=1 go build -ldflags "-s -w" -o lilurl .
```

### Database Migrations

Migrations are managed using dbmate and stored in `schema/migrations/`:

```bash
# Create a new migration
dbmate new create_users_table

# Apply migrations
make migrate

# Rollback migrations
dbmate rollback
```

## 🧪 Testing

```bash
# Run all tests
go test ./...

# Run tests with coverage
go test -cover ./...

# Run tests with race detector
go test -race ./...

# Run specific test
go test -v ./config -run TestConfig
```

## 🚀 Deployment

### Docker Deployment

```bash
# Build and run with docker-compose
docker-compose up --build

# Run in detached mode
docker-compose up -d

# View logs
docker-compose logs -f lilurl
```

### Production Deployment (Google Cloud)

```bash
# Build for Google Artifact Registry
make build-ar GAR=<registry-url>

# Push to registry
make push-ar GAR=<registry-url>
```

### Nginx Proxy Configuration

The project includes an Nginx reverse proxy configuration in `proxy/nginx.conf` for production deployments.

## 📁 Project Structure

```
lilurl/
├── app/                        # Application layer
│   ├── handlers/              # HTTP handlers
│   │   ├── routes.go         # Route definitions
│   │   └── lilurl/           # LilURL handler implementation
│   └── models/               # Data models and DB logic
├── config/                    # Configuration management
│   ├── config.go             # Config struct and loader
│   └── config_test.go        # Config tests
├── internal/                  # Private packages
│   └── pkg/
│       └── generator/        # URL shortening algorithms
├── pkg/                       # Public packages
│   ├── database/             # Database utilities
│   │   └── sqlite/           # SQLite initialization
│   └── log/                  # Logging setup
├── schema/                    # Database schema
│   ├── migrations/           # Database migrations
│   └── lilurl.db            # SQLite database file
├── proxy/                     # Proxy configuration
│   └── nginx.conf           # Nginx configuration
├── main.go                    # Application entry point
├── Dockerfile                 # Docker configuration
├── docker-compose.yaml        # Docker Compose setup
├── Makefile                   # Build automation
├── go.mod                     # Go module definition
├── go.sum                     # Go module checksums
├── config.yaml                # Default configuration
└── README.md                  # This file
```

## 🤝 Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add some amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

### Development Guidelines

- Follow Go best practices and idioms
- Ensure all tests pass before submitting PR
- Add tests for new features
- Update documentation as needed
- Use conventional commit messages

## 📄 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## 🙏 Acknowledgments

- [Fiber](https://github.com/gofiber/fiber) - The web framework used
- [SQLite](https://www.sqlite.org/) - The database engine
- [sqlx](https://github.com/jmoiron/sqlx) - Database toolkit
- [cleanenv](https://github.com/ilyakaznacheev/cleanenv) - Configuration management

---

<p align="center">Made with ❤️ by <a href="https://github.com/pansachin">@pansachin</a></p>
