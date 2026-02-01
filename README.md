# Search Engine Service

A Go/Gin-based search engine service that aggregates content from multiple providers, calculates relevance scores, and provides a RESTful API with a dashboard interface.

## Overview

This project demonstrates enterprise-grade software architecture principles and design patterns. It aggregates content from multiple external providers (JSON/XML), calculates sophisticated relevance scores, and provides search capabilities through a RESTful API and web dashboard.

## Architecture

This project follows **Clean Architecture** principles with clear separation of concerns and dependency inversion:

```
┌─────────────────────────────────────────────────────────┐
│                    HTTP Layer (Gin)                      │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐  │
│  │   Handlers   │  │  Middleware  │  │  Templates   │  │
│  └──────────────┘  └──────────────┘  └──────────────┘  │
└─────────────────────────────────────────────────────────┘
                          │
                          ▼
┌─────────────────────────────────────────────────────────┐
│                  Service Layer                          │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐  │
│  │   Content    │  │   Provider   │  │   Scoring    │  │
│  │   Service    │  │   Service    │  │   Service    │  │
│  └──────────────┘  └──────────────┘  └──────────────┘  │
└─────────────────────────────────────────────────────────┘
         │                    │                  │
         ▼                    ▼                  ▼
┌──────────────┐  ┌──────────────────┐  ┌──────────────┐
│  Repository  │  │  Adapter Pattern │  │   Domain     │
│   Layer      │  │  (Providers)     │  │   Models     │
└──────────────┘  └──────────────────┘  └──────────────┘
         │                    │
         ▼                    ▼
┌──────────────┐  ┌──────────────────┐
│  PostgreSQL  │  │  External APIs    │
│  Database    │  │  (JSON/XML)      │
└──────────────┘  └──────────────────┘
```

### Project Structure

```
search-engine-go/
├── cmd/                    # Application entry points
│   └── api/               # Main API server
├── internal/              # Private application code
│   ├── api/              # HTTP handlers and middleware
│   │   ├── handler/      # Request handlers
│   │   └── middleware/   # HTTP middleware
│   ├── config/           # Configuration management
│   ├── domain/           # Business domain models and specifications
│   ├── infrastructure/   # External dependencies
│   │   ├── cache/        # Cache implementations (Redis, In-Memory)
│   │   ├── database/     # Database connection and migrations
│   │   └── logger/       # Logging infrastructure
│   ├── repository/       # Data access layer
│   └── service/          # Business logic layer
├── pkg/                   # Public reusable packages
│   └── adapter/          # Provider adapter pattern implementation
├── web/                   # Frontend assets
│   ├── static/           # Static files (CSS, JS)
│   └── templates/        # HTML templates
└── migrations/            # Database migrations
```

## Design Patterns

### 1. Clean Architecture

**Implementation:**

- **Domain Layer** (`internal/domain/`): Core business entities and rules, no external dependencies
- **Service Layer** (`internal/service/`): Business logic orchestration
- **Repository Layer** (`internal/repository/`): Data access abstraction
- **Infrastructure Layer** (`internal/infrastructure/`): External dependencies (database, cache, logging)
- **API Layer** (`internal/api/`): HTTP handlers and middleware

**Benefits:**

- **Dependency Rule**: Dependencies point inward (infrastructure → service → domain)
- **Testability**: Business logic independent of frameworks
- **Maintainability**: Changes to external dependencies don't affect core logic
- **Flexibility**: Easy to swap implementations (e.g., Redis ↔ In-Memory cache)

### 2. SOLID Principles

#### Single Responsibility Principle (SRP)

- **ContentService**: Orchestrates content operations only
- **ProviderService**: Manages provider fetching only
- **ScoringService**: Calculates scores only
- **ContentRepository**: Handles data access only
- Each handler focuses on a single HTTP concern

#### Open/Closed Principle (OCP)

- **Adapter Pattern**: New providers can be added without modifying existing code
- **Specification Pattern**: New scoring rules can be added by implementing `ScoreSpecification` interface
- **Cache Interface**: Can swap Redis/In-Memory without changing service code

#### Liskov Substitution Principle (LSP)

- All `ProviderAdapter` implementations are interchangeable
- `ScoreSpecification` implementations can be swapped
- Cache implementations (Redis/In-Memory) are substitutable

#### Interface Segregation Principle (ISP)

- Focused interfaces: `ProviderAdapter`, `ScoreSpecification`, `Cache`
- Clients depend only on methods they use
- No fat interfaces forcing unnecessary dependencies

#### Dependency Inversion Principle (DIP)

- Services depend on abstractions (interfaces), not concrete implementations
- Dependencies injected through constructors
- High-level modules (services) don't depend on low-level modules (infrastructure)

### 3. Specification Pattern

**Purpose**: Encapsulate business rules and criteria in reusable, composable specifications.

**Implementation:**

#### Score Specifications (`internal/domain/score_specification.go`)

- **`ScoreSpecification` Interface**: Defines contract for score calculations
- **`ContentPopularityScoreSpecification`**: Calculates base popularity score
- **`VideoTypeBoostSpecification`**: Applies video content boost
- **`RecentContentBoostSpecification`**: Applies freshness boost based on content age
- **`ContentQualityRatioSpecification`**: Calculates engagement ratio
- **`CompositeScoreSpecification`**: Combines multiple specifications
- **`ContentRelevanceScoreSpecification`**: Final composite specification

**Benefits:**

- **Composability**: Specifications can be combined (Composite Pattern)
- **Testability**: Each specification tested independently
- **Flexibility**: Easy to add/modify scoring rules
- **Domain-Driven Design**: Business rules live in domain layer

#### Pagination Specification (`internal/domain/pagination_specification.go`)

- Encapsulates pagination validation and normalization logic
- Ensures consistent pagination behavior across handlers

### 4. Adapter Pattern

**Purpose**: Integrate multiple providers with different formats (JSON/XML) through a unified interface.

**Implementation:**

- **`ProviderAdapter` Interface**: Standard contract for all providers
- **`AdapterRegistry`**: Manages provider registration and retrieval
- **`JSONProviderAdapter`**: Adapts JSON format providers
- **`XMLProviderAdapter`**: Adapts XML format providers

**Benefits:**

- **Extensibility**: Add new providers without modifying existing code
- **Polymorphism**: Services work with `ProviderAdapter` interface
- **Testability**: Easy to mock providers for testing
- **Separation of Concerns**: Provider-specific logic isolated

### 5. Repository Pattern

**Purpose**: Abstract data access and provide a clean interface for domain operations.

**Implementation:**

- **`ContentRepository`**: Provides methods for content persistence and retrieval
- Hides GORM/database implementation details
- Supports batch operations and transactions

**Benefits:**

- **Testability**: Easy to mock for unit testing
- **Flexibility**: Can swap database implementations
- **Maintainability**: Database changes isolated to repository
- **Domain Focus**: Services work with domain models, not database models

### 6. Service Layer Pattern

**Purpose**: Encapsulate business logic and coordinate between repositories and adapters.

**Implementation:**

- **`ContentService`**: Orchestrates content search, caching, and persistence
- **`ProviderService`**: Manages concurrent provider fetching with rate limiting
- **`ScoringService`**: Delegates to score specifications

**Benefits:**

- **Separation of Concerns**: Business logic separated from HTTP and data access
- **Reusability**: Services can be used by multiple handlers
- **Testability**: Business logic tested independently
- **Transaction Management**: Coordinates multi-step operations

### 7. Dependency Injection

**Purpose**: Achieve loose coupling and improve testability.

**Implementation:**

- All dependencies injected through constructors
- No global state or singletons
- Composition root in `cmd/api/main.go` wires all dependencies

**Example:**

```go
func NewContentService(
    repo *repository.ContentRepository,
    providerSvc *ProviderService,
    scoringSvc *ScoringService,
    cache cache.Cache,
    log *zap.Logger,
) *ContentService
```

**Benefits:**

- **Testability**: Easy to inject mocks
- **Flexibility**: Easy to swap implementations
- **Maintainability**: Clear dependency graph
- **No Hidden Dependencies**: All dependencies explicit

### 8. Composite Pattern

**Purpose**: Compose multiple specifications into a single specification.

**Implementation:**

- **`CompositeScoreSpecification`**: Combines multiple `ScoreSpecification` implementations
- Used in `ContentRelevanceScoreSpecification` to combine popularity, freshness, and engagement scores

**Benefits:**

- **Flexibility**: Mix and match specifications
- **Reusability**: Specifications can be reused in different combinations
- **Maintainability**: Each specification remains focused

## Key Features

- **Provider Integration**: Adapter pattern for easy integration of new providers (JSON/XML)
- **Content Scoring**: Sophisticated scoring algorithm using Specification Pattern
- **Search & Filtering**: Full-text search with content type filtering and sorting
- **Caching**: Multi-layer caching (Redis/In-Memory) with interface-based abstraction
- **Rate Limiting**: Per-provider rate limiting with exponential backoff
- **Dashboard**: Simple web interface for content browsing
- **PostgreSQL with GORM**: Robust data persistence with ORM-based operations
- **Clean Architecture**: Maintainable and testable code structure

## Technology Stack

### Backend

- **Go 1.21+**: Excellent performance and concurrency support
- **Gin**: Fast HTTP web framework
- **GORM**: Feature-rich ORM for database operations
- **PostgreSQL**: Robust relational database with full-text search

### Infrastructure

- **Redis**: Distributed caching (with In-Memory fallback)
- **Zap**: High-performance structured logging

## Scoring Algorithm

The content scoring algorithm uses the **Specification Pattern** to calculate relevance:

```
Final Score = (Base Score × Type Coefficient) + Freshness Score + Engagement Score
```

**Components:**

1. **Base Score** (`ContentPopularityScoreSpecification`):

   - Video: `views / 1000 + likes / 100`
   - Text: `reading_time + reactions / 50`

2. **Type Coefficient** (`VideoTypeBoostSpecification`):

   - Video: 1.5
   - Text: 1.0

3. **Freshness Score** (`RecentContentBoostSpecification`):

   - 1 week: +5
   - 1 month: +3
   - 3 months: +1
   - Older: +0

4. **Engagement Score** (`ContentQualityRatioSpecification`):
   - Video: `(likes / views) × 10`
   - Text: `(reactions / reading_time) × 5`

All specifications are composed using `CompositeScoreSpecification` in `ContentRelevanceScoreSpecification`.

## API Endpoints

### Search Content

```
GET /api/v1/search?query=<keyword>&content_type=<video|text>&page=<page>&page_size=<size>&sort_by=<score|created_at|popularity>
```

### Get Content by ID

```
GET /api/v1/content/:id
```

### Dashboard

```
GET /dashboard
GET /
```

### Health Check

```
GET /health
```

## Installation and Running

### Prerequisites

- **Go**: 1.21 or higher
- **PostgreSQL**: 12 or higher (or Docker)
- **Redis**: 6.0 or higher (optional, in-memory cache fallback available)
- **Make**: For build commands (optional)

### Quick Start

#### 1. Clone the Project

```bash
git clone <repository-url>
cd search-engine-go
```

#### 2. Install Dependencies

```bash
go mod download
go mod tidy
```

#### 3. Set Up Environment Variables

```bash
cp .env.example .env
```

Edit the `.env` file with your settings:

```bash
# Database settings
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=postgres
DB_NAME=search_engine

# Server settings
SERVER_PORT=8080

# Cache settings (to use Redis)
CACHE_TYPE=redis
CACHE_HOST=localhost
CACHE_PORT=6379
```

#### 4. Set Up Database

**Option A: PostgreSQL with Docker**

```bash
docker run -d \
  --name search-engine-postgres \
  -e POSTGRES_USER=postgres \
  -e POSTGRES_PASSWORD=postgres \
  -e POSTGRES_DB=search_engine \
  -p 5432:5432 \
  postgres:15-alpine
```

**Option B: Local PostgreSQL**

```bash
createdb search_engine
```

#### 5. Run Database Migrations

```bash
psql -h localhost -U postgres -d search_engine \
  -f internal/infrastructure/database/migrations/001_initial_schema.up.sql
```

#### 6. Start Redis (Optional)

```bash
docker run -d \
  --name search-engine-redis \
  -p 6379:6379 \
  redis:7-alpine
```

If you don't want to use Redis, set `CACHE_TYPE=memory` in the `.env` file.

#### 7. Run the Application

**Using Make:**

```bash
make run
```

**Or directly:**

```bash
go run cmd/api/main.go
```

The application will run at `http://localhost:8080`.

### Installation with Docker Compose

To start all services with a single command:

```bash
docker-compose up -d
```

This command starts:

- PostgreSQL (port 5432)
- Redis (port 6379)
- API Server (port 8080)

To run migrations:

```bash
docker exec -i search-engine-postgres psql -U postgres -d search_engine \
  < internal/infrastructure/database/migrations/001_initial_schema.up.sql
```

### Testing the Application

#### Health Check

```bash
curl http://localhost:8080/health
```

Expected response:

```json
{ "status": "ok" }
```

#### Login and Get Token

```bash
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username":"test","password":"test"}'
```

#### Search Content

```bash
curl -X GET "http://localhost:8080/api/v1/search?query=golang&page=1&page_size=20" \
  -H "Authorization: Bearer YOUR_TOKEN_HERE"
```

#### Access Dashboard

Open the following address in your browser:

```
http://localhost:8080/dashboard
```

### Development Environment Setup

#### Install Development Tools

```bash
make install-tools
```

This command installs:

- `golangci-lint`: Code quality checks
- `goimports`: Import formatting

#### Code Formatting

```bash
make fmt
```

#### Lint Check

```bash
make lint
```

#### Running Tests

```bash
make test              # All tests
make test-unit         # Unit tests only
make test-integration  # Integration tests only
```

#### Build Binary

```bash
make build
```

The binary file is created at `bin/api`.

### Production Deployment

#### Build Docker Image

```bash
make docker-build
```

#### Run Docker Container

```bash
make docker-run
```

#### Environment Variables

In production environment, make sure to set the following variables:

```bash
ENVIRONMENT=production
JWT_SECRET=<strong-secret-key>
DB_PASSWORD=<secure-password>
CACHE_PASSWORD=<redis-password>
```

### Troubleshooting

#### Database Connection Error

**Error:**

```
Failed to connect to database: connection refused
```

**Solution:**

- Ensure PostgreSQL is running: `docker ps` or `ps aux | grep postgres`
- Check database information in `.env` file
- Verify database exists: `psql -U postgres -l`

#### Port Already in Use

**Error:**

```
bind: address already in use
```

**Solution:**

- Change `SERVER_PORT` value in `.env` file
- Or stop the process using the port:
  ```bash
  lsof -ti:8080 | xargs kill
  ```

#### Cache Connection Error

If the application cannot connect to Redis, it automatically falls back to in-memory cache. This is normal for development environments.

#### Provider Connection Error

Providers are optional. If provider URLs are not accessible, the application continues to run but cannot fetch new content.

### Configuration Reference

See `.env.example` file for all configuration options. Important parameters:

- **SERVER_PORT**: API server port (default: 8080)
- **DB\_\***: PostgreSQL connection information
- **CACHE_TYPE**: `redis` or `memory`
- **PROVIDER1_URL, PROVIDER2_URL**: Provider endpoints
- **LOG_LEVEL**: `debug`, `info`, `warn`, `error`
- **JWT_SECRET**: Secret key for JWT token signing
- **JWT_EXPIRATION**: Token validity duration (e.g., `24h`)

## Documentation

- **API Documentation**: [docs/API.md](./docs/API.md) - Detailed API endpoint documentation
- **Quick Start**: [docs/QUICKSTART.md](./docs/QUICKSTART.md) - Quick installation guide
- **Technology Choices**: [docs/TECHNOLOGY_CHOICES.md](./docs/TECHNOLOGY_CHOICES.md) - Technology selection justifications
- **Non-Functional Requirements**: [docs/NON_FUNCTIONAL_REQUIREMENTS.md](./docs/NON_FUNCTIONAL_REQUIREMENTS.md) - Performance and security requirements

## Development

### Running Tests

```bash
make test              # Run all tests
make test-unit         # Run unit tests only
make test-integration  # Run integration tests only
```

### Code Quality

```bash
make lint  # Run linter
make fmt   # Format code
```

## Design Principles Summary

This project demonstrates:

1. **Clean Architecture**: Clear separation of concerns with dependency inversion
2. **SOLID Principles**: Applied throughout the codebase
3. **Specification Pattern**: Encapsulates business rules in composable specifications
4. **Adapter Pattern**: Enables easy provider integration
5. **Repository Pattern**: Abstracts data access
6. **Service Layer**: Encapsulates business logic
7. **Dependency Injection**: Achieves loose coupling
8. **Composite Pattern**: Combines specifications flexibly

All patterns work together to create a maintainable, testable, and extensible codebase that follows enterprise software development best practices.
