# Production API

A production-ready HTTP API server built with Go, featuring middleware for logging, rate limiting, CORS, compression, request tracing, panic recovery, and timeout handling. Includes RESTful endpoints for health checks and user management.

---

## Prerequisites

| Requirement | Version |
|-------------|---------|
| Go | 1.21 or higher |
| Operating System | Linux, macOS, Windows |

---

## Installation

```bash
# Clone the repository
git clone https://github.com/yourusername/production-api.git
cd production-api

# Download dependencies
go mod download

# Build the binary
go build -o api ./cmd/api

# Run the server
./api
```

---

## Configuration

The server is configured using environment variables:

| Variable | Default | Description |
|----------|---------|-------------|
| `PORT` | `8080` | Server listening port |
| `RATE_LIMIT` | `100` | Maximum requests per IP per window |
| `RATE_WINDOW_SECONDS` | `60` | Rate limit time window in seconds |
| `TIMEOUT_SECONDS` | `30` | Request timeout in seconds |
| `ALLOWED_ORIGINS` | `*` | Comma-separated list of allowed CORS origins |

### Example Configuration

```bash
export PORT=8081
export RATE_LIMIT=50
export RATE_WINDOW_SECONDS=30
export TIMEOUT_SECONDS=10
export ALLOWED_ORIGINS=https://example.com,https://app.com

go run cmd/api/main.go
```

---

## API Endpoints

### Health Check

```
GET /health
```

**Response:**
```json
{
  "status": "ok",
  "timestamp": "2026-05-31T10:30:00Z"
}
```

---

### Get All Users

```
GET /users
```

**Response:**
```json
[
  {
    "id": "1",
    "name": "Alice",
    "email": "alice@example.com"
  },
  {
    "id": "2",
    "name": "Bob",
    "email": "bob@example.com"
  },
  {
    "id": "3",
    "name": "Charlie",
    "email": "charlie@example.com"
  }
]
```

---

### Get User by ID

```
GET /users/{id}
```

**Response (200 OK):**
```json
{
  "id": "1",
  "name": "Alice",
  "email": "alice@example.com"
}
```

**Response (404 Not Found):**
```json
{
  "error": "user not found"
}
```

---

### Create User

```
POST /users
Content-Type: application/json
```

**Request Body:**
```json
{
  "name": "David",
  "email": "david@example.com"
}
```

**Response (201 Created):**
```json
{
  "id": "4",
  "name": "David",
  "email": "david@example.com"
}
```

---

## Run Instructions

### Development Mode

```bash
# Run directly with go
go run cmd/api/main.go

# With custom port
PORT=8081 go run cmd/api/main.go

# With custom rate limit
RATE_LIMIT=50 go run cmd/api/main.go
```

### Production Mode

```bash
# Build the binary
go build -o api ./cmd/api

# Run the binary
./api

# Or with environment variables
PORT=8081 RATE_LIMIT=100 ./api
```

### Using systemd (Linux)

Create `/etc/systemd/system/production-api.service`:

```ini
[Unit]
Description=Production API Server
After=network.target

[Service]
Type=simple
User=www-data
WorkingDirectory=/opt/production-api
ExecStart=/opt/production-api/api
Restart=always
Environment="PORT=8080"
Environment="RATE_LIMIT=100"

[Install]
WantedBy=multi-user.target
```

Then run:

```bash
sudo systemctl enable production-api
sudo systemctl start production-api
```

---

## Testing Instructions

### Run All Tests

```bash
# Run all tests
go test ./...

# Run with verbose output
go test -v ./...

# Run with coverage report
go test -cover ./...
```

### Run Specific Test Packages

```bash
# Test only middleware
go test ./internal/middleware/ -v

# Test only handlers
go test ./internal/handlers/ -v

# Test only config
go test ./internal/config/ -v
```

### Run Specific Test

```bash
# Run a single test function
go test ./internal/handlers/ -run TestCreateUser -v

# Run tests matching pattern
go test ./internal/middleware/ -run "TestRateLimit" -v
```

### Test with Race Detection

```bash
go test -race ./...
```

### Manual Endpoint Testing

```bash
# Health check
curl http://localhost:8080/health

# Get all users
curl http://localhost:8080/users

# Get user by ID
curl http://localhost:8080/users/1

# Create user
curl -X POST http://localhost:8080/users \
  -H "Content-Type: application/json" \
  -d '{"name":"Test","email":"test@example.com"}'

# Test rate limiting
for i in {1..10}; do
  curl -s -o /dev/null -w "Request $i: %{http_code}\n" http://localhost:8080/users
done

# Test with custom request ID
curl -i -H "X-Request-ID: my-custom-123" http://localhost:8080/health

# Test compression
curl -H "Accept-Encoding: gzip" http://localhost:8080/users --compressed -i

# Test CORS preflight
curl -i -X OPTIONS http://localhost:8080/users \
  -H "Origin: https://example.com" \
  -H "Access-Control-Request-Method: POST"
```

---

## Middleware Features

| Feature | Description |
|---------|-------------|
| Request ID | Auto-generated UUID for request tracing |
| Rate Limiting | Per-IP rate limiting with configurable window |
| CORS | Cross-origin request support with configurable origins |
| Compression | Gzip compression for responses >512 bytes |
| Logging | Structured JSON logging with slog |
| Recovery | Panic recovery with stack trace logging |
| Timeout | Configurable request timeouts (504 on timeout) |
| Real IP | Extracts client IP from headers |

---

## Project Structure

```
production-api/
├── cmd/
│   └── api/
│       └── main.go
├── internal/
│   ├── config/
│   │   ├── config.go
│   │   └── config_test.go
│   ├── handlers/
│   │   ├── health.go
│   │   ├── health_test.go
│   │   ├── users.go
│   │   └── users_test.go
│   └── middleware/
│       ├── compress.go
│       ├── compress_test.go
│       ├── cors.go
│       ├── cors_test.go
│       ├── logger.go
│       ├── logger_test.go
│       ├── ratelimit.go
│       ├── ratelimit_test.go
│       ├── realip.go
│       ├── realip_test.go
│       ├── recovery.go
│       ├── recovery_test.go
│       ├── requestid.go
│       ├── requestid_test.go
│       ├── timeout.go
│       └── timeout_test.go
├── go.mod
├── go.sum
└── README.md
```

---

## License

MIT