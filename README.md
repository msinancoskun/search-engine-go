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

## Quick Start

See [QUICKSTART.md](./docs/QUICKSTART.md) for detailed setup and installation instructions.

## Documentation

- **Architecture Details**: See [docs/ARCHITECTURE.md](./docs/ARCHITECTURE.md)
- **Non-Functional Requirements**: See [NON_FUNCTIONAL_REQUIREMENTS.md](./NON_FUNCTIONAL_REQUIREMENTS.md)

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
