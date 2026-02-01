# Quick Start Guide

Get up and running with the Search Engine Service in minutes.

## Prerequisites Check

```bash
# Check Go version (requires 1.21+)
go version

# Check PostgreSQL (optional - can use Docker)
psql --version

# Check Docker (optional - for containerized setup)
docker --version
```

## Option 1: Local Development Setup

### 1. Install Dependencies

```bash
go mod download
go mod tidy
```

### 2. Set Up Environment

```bash
cp .env.example .env
# Edit .env with your settings (or use defaults)
```

### 3. Set Up Database

**Using PostgreSQL directly:**
```bash
createdb search_engine
psql -U postgres -d search_engine -f internal/infrastructure/database/migrations/001_initial_schema.up.sql
```

**Using Docker:**
```bash
docker run -d \
  --name search-engine-postgres \
  -e POSTGRES_USER=postgres \
  -e POSTGRES_PASSWORD=postgres \
  -e POSTGRES_DB=search_engine \
  -p 5432:5432 \
  postgres:15-alpine

# Wait a few seconds for PostgreSQL to start, then run migrations
psql -h localhost -U postgres -d search_engine -f internal/infrastructure/database/migrations/001_initial_schema.up.sql
```

### 4. Run the Application

```bash
# Using Make
make run

# Or directly
go run cmd/api/main.go
```

The API will be available at `http://localhost:8080`

## Option 2: Docker Compose Setup

### 1. Start All Services

```bash
docker-compose up -d
```

This starts:
- PostgreSQL (port 5432)
- Redis (port 6379)
- API Server (port 8080)

### 2. Run Migrations

```bash
# Get into the postgres container
docker exec -it search-engine-postgres psql -U postgres -d search_engine

# Or run migrations from host
docker exec -i search-engine-postgres psql -U postgres -d search_engine < internal/infrastructure/database/migrations/001_initial_schema.up.sql
```

### 3. Access the Application

- API: `http://localhost:8080`
- Dashboard: `http://localhost:8080/dashboard`
- Health Check: `http://localhost:8080/health`

## Testing the API

### Search Content

```bash
curl "http://localhost:8080/api/v1/search?query=test&page=1&page_size=20"
```

### Get Content by ID

```bash
curl "http://localhost:8080/api/v1/content/1"
```

### Health Check

```bash
curl "http://localhost:8080/health"
```

## Setting Up Mock Providers

The application expects providers at:
- Provider 1 (JSON): `http://localhost:3001/api/content`
- Provider 2 (XML): `http://localhost:3002/api/content`

You can:
1. Update `.env` to point to your actual provider URLs
2. Create mock providers (see examples below)
3. Test with the database directly (providers are optional for testing)

## Common Issues

### Database Connection Error

```
Failed to connect to database: connection refused
```

**Solution:**
- Ensure PostgreSQL is running
- Check connection details in `.env`
- Verify database exists: `psql -U postgres -l`

### Port Already in Use

```
bind: address already in use
```

**Solution:**
- Change `SERVER_PORT` in `.env`
- Or stop the process using port 8080

### Cache Connection Error

The application will automatically fall back to in-memory cache if Redis is unavailable. This is expected behavior for development.

## Next Steps

1. **Review the Architecture**: See [PROJECT_STRUCTURE.md](./PROJECT_STRUCTURE.md)
2. **Read Non-Functional Requirements**: See [NON_FUNCTIONAL_REQUIREMENTS.md](./NON_FUNCTIONAL_REQUIREMENTS.md)
3. **Add Providers**: See [pkg/adapter/README.md](./pkg/adapter/README.md)
4. **Run Tests**: `make test`

## Development Workflow

```bash
# Format code
make fmt

# Run linter
make lint

# Run tests
make test

# Build binary
make build

# Run application
make run
```

## Environment Variables Reference

Key environment variables (see `.env.example` for full list):

- `SERVER_PORT`: API server port (default: 8080)
- `DB_HOST`: PostgreSQL host (default: localhost)
- `DB_NAME`: Database name (default: search_engine)
- `CACHE_TYPE`: "redis" or "memory" (default: memory)
- `PROVIDER1_URL`: First provider endpoint
- `PROVIDER2_URL`: Second provider endpoint
- `LOG_LEVEL`: "debug", "info", "warn", "error" (default: info)
