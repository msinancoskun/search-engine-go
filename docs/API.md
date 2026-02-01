# API Documentation

This documentation covers all endpoints of the Search Engine Service API, including request/response formats and usage examples.

## Table of Contents

- [General Information](#general-information)
- [Authentication](#authentication)
- [Endpoints](#endpoints)
  - [Search](#search)
  - [Content Details](#content-details)
  - [Health Check](#health-check)
  - [Dashboard](#dashboard)
- [Error Handling](#error-handling)
- [Rate Limiting](#rate-limiting)
- [Examples](#examples)

## General Information

### Base URL

- **Development**: `http://localhost:8080`
- **Production**: `https://api.example.com`

### API Version

The API currently runs on **v1**. All endpoints start with the `/api/v1/` prefix.

### Content-Type

All requests and responses must be in `application/json` format.

### Response Format

Successful responses return with HTTP 200 status code. Error cases return the appropriate HTTP status code and error message.

## Authentication

The API uses JWT (JSON Web Token) based authentication. **All endpoints except login and health check require authentication.**

### Login Endpoint

**POST** `/api/v1/auth/login`

Authenticate user and obtain a JWT token. This is the only public authentication endpoint.

#### Request Body

```json
{
  "username": "username",
  "password": "password"
}
```

#### Response (200 OK)

```json
{
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
}
```

**Note:** The token expires after 24 hours (configurable via `JWT_EXPIRATION` environment variable).

#### Response (401 Unauthorized)

```json
{
  "error": "Invalid credentials"
}
```

### Logout Endpoint

**POST** `/api/v1/auth/logout`

Logout and invalidate the current session. Requires authentication.

#### Headers

```
Authorization: Bearer <token>
```

#### Response (200 OK)

```json
{
  "message": "Logged out successfully"
}
```

**Note:** Since JWT tokens are stateless, logout is primarily handled client-side by removing the token. This endpoint is provided for consistency and future enhancements.

### Token Usage

The obtained token must be sent in the `Authorization` header when making requests to protected endpoints:

```
Authorization: Bearer <token>
```

**Example:**

```bash
curl -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..." \
  http://localhost:8080/api/v1/search?query=golang
```

## Endpoints

### Search

**GET** `/api/v1/search`

Used to search for content. Aggregates content from multiple providers, calculates scores, and returns results.

#### Query Parameters

| Parameter      | Type    | Required | Default | Description                                        |
| -------------- | ------- | -------- | ------- | -------------------------------------------------- |
| `query`        | string  | No       | -       | Search query (searches in title and content)       |
| `content_type` | enum    | No       | -       | Content type filter: `video` or `text`             |
| `page`         | integer | No       | 1       | Page number (starts from 1)                        |
| `page_size`    | integer | No       | 20      | Number of records per page (between 1-100)         |
| `sort_by`      | enum    | No       | `score` | Sort criteria: `score`, `created_at`, `popularity` |
| `sort_order`   | enum    | No       | `desc`  | Sort order: `asc` or `desc`                        |

#### Example Request

```bash
GET /api/v1/search?query=golang&content_type=video&page=1&page_size=20&sort_by=score&sort_order=desc
```

#### Response (200 OK)

```json
{
  "items": [
    {
      "id": 1,
      "provider_id": "provider1_123",
      "provider": "provider1",
      "title": "Go Programming Tutorial",
      "type": "video",
      "views": 10000,
      "likes": 500,
      "reading_time": 0,
      "reactions": 0,
      "score": 25.5,
      "created_at": "2024-01-15T10:30:00Z"
    },
    {
      "id": 2,
      "provider_id": "provider2_456",
      "provider": "provider2",
      "title": "Advanced Go Patterns",
      "type": "text",
      "views": 0,
      "likes": 0,
      "reading_time": 15,
      "reactions": 120,
      "score": 18.2,
      "created_at": "2024-01-10T08:20:00Z"
    }
  ],
  "total": 42,
  "page": 1,
  "page_size": 20,
  "total_pages": 3
}
```

#### Response Fields

- `items`: List of found content items
- `total`: Total number of matching records
- `page`: Current page number
- `page_size`: Number of records per page
- `total_pages`: Total number of pages

#### Content Object Fields

- `id`: Content ID (unique)
- `provider_id`: Provider-specific content ID
- `provider`: Provider name
- `title`: Content title
- `type`: Content type (`video` or `text`)
- `views`: Number of views (for video content)
- `likes`: Number of likes
- `reading_time`: Reading time in minutes (for text content)
- `reactions`: Number of reactions (for text content)
- `score`: Calculated relevance score
- `created_at`: Creation date (in ISO 8601 format)

#### Error Cases

**400 Bad Request** - Invalid parameters

```json
{
  "error": "Invalid input",
  "code": "invalid_input",
  "details": {
    "field": "page_size",
    "reason": "must be between 1 and 100"
  },
  "request_id": "abc123"
}
```

**401 Unauthorized** - Missing or invalid token

```json
{
  "error": "Authentication required",
  "message": "Authorization header is required. Please provide a valid JWT token."
}
```

Or for invalid/expired tokens:

```json
{
  "error": "Invalid or expired token",
  "message": "The provided token is invalid, expired, or malformed. Please login again."
}
```

**429 Too Many Requests** - Rate limit exceeded

```json
{
  "error": "Rate limit exceeded",
  "message": "Too many requests. Please try again later.",
  "request_id": "abc123"
}
```

**500 Internal Server Error** - Server error

```json
{
  "error": "Internal server error",
  "request_id": "abc123"
}
```

### Content Details

**GET** `/api/v1/content/:id`

Retrieves details of a specific content item.

#### Path Parameters

| Parameter | Type    | Required | Description |
| --------- | ------- | -------- | ----------- |
| `id`      | integer | Yes      | Content ID  |

#### Example Request

```bash
GET /api/v1/content/1
```

#### Response (200 OK)

```json
{
  "id": 1,
  "provider_id": "provider1_123",
  "provider": "provider1",
  "title": "Go Programming Tutorial",
  "type": "video",
  "views": 10000,
  "likes": 500,
  "reading_time": 0,
  "reactions": 0,
  "score": 25.5,
  "created_at": "2024-01-15T10:30:00Z"
}
```

#### Error Cases

**400 Bad Request** - Invalid ID format

```json
{
  "error": "Invalid input",
  "code": "invalid_input",
  "details": {
    "field": "id",
    "reason": "must be a valid integer"
  },
  "request_id": "abc123"
}
```

**404 Not Found** - Content not found

```json
{
  "error": "Content not found",
  "code": "not_found",
  "details": {
    "resource": "content",
    "id": 999
  },
  "request_id": "abc123"
}
```

**401 Unauthorized** - Missing or invalid token

```json
{
  "error": "Authentication required",
  "message": "Authorization header is required. Please provide a valid JWT token."
}
```

### Health Check

**GET** `/health`

Checks if the API is running and healthy. This is the only public endpoint (along with login). Does not require authentication.

#### Example Request

```bash
GET /health
```

#### Response (200 OK)

```json
{
  "status": "ok"
}
```

### Dashboard

**GET** `/dashboard` or **GET** `/`

Provides access to the web-based dashboard interface. **Requires JWT authentication.** Users will be redirected to `/login` if not authenticated.

#### Features

- Content search and filtering
- Pagination support
- Content detail viewing
- Responsive design

## Error Handling

### Error Response Format

All error responses return in the following format:

```json
{
  "error": "Error type",
  "message": "Detailed error message (optional)",
  "request_id": "Request ID for tracking (optional)"
}
```

Some errors may also include additional fields like `code` and `details` for validation errors.

### HTTP Status Codes

| Status Code | Description                |
| ----------- | -------------------------- |
| 200         | Successful request         |
| 400         | Invalid request parameters |
| 401         | Authentication error       |
| 404         | Resource not found         |
| 429         | Rate limit exceeded        |
| 500         | Server error               |

### Error Types

#### Validation Errors (400)

When invalid parameters are sent:

```json
{
  "error": "Invalid input",
  "code": "invalid_input",
  "details": {
    "field": "page",
    "reason": "must be a positive integer"
  },
  "request_id": "abc123"
}
```

#### Authentication Errors (401)

When token is missing:

```json
{
  "error": "Authentication required",
  "message": "Authorization header is required. Please provide a valid JWT token."
}
```

When token is invalid or expired:

```json
{
  "error": "Invalid or expired token",
  "message": "The provided token is invalid, expired, or malformed. Please login again."
}
```

When authorization header format is invalid:

```json
{
  "error": "Invalid authorization header",
  "message": "Authorization header must be in the format: 'Bearer <token>'"
}
```

#### Not Found Errors (404)

When the requested resource is not found:

```json
{
  "error": "Content not found",
  "code": "not_found",
  "details": {
    "resource": "content",
    "id": 999
  },
  "request_id": "abc123"
}
```

#### Rate Limit Errors (429)

When rate limit is exceeded:

```json
{
  "error": "Rate limit exceeded",
  "message": "Too many requests. Please try again later.",
  "request_id": "abc123"
}
```

#### Server Errors (500)

For unexpected server errors:

```json
{
  "error": "Internal server error",
  "request_id": "abc123"
}
```

## Rate Limiting

The API implements rate limiting to prevent abuse.

### Limits

- **Default Limit**: 100 requests/minute/IP address
- **Limit Exceeded**: Returns 429 Too Many Requests error

### Rate Limit Exceeded

When rate limit is exceeded:

```json
{
  "error": "Rate limit exceeded",
  "message": "Too many requests. Please try again later.",
  "request_id": "abc123"
}
```

**Note:** Rate limiting applies only to API endpoints (`/api/*`). HTML pages (`/`, `/dashboard`, `/login`, `/docs/*`) and static assets are excluded from rate limiting.

## Examples

### cURL Examples

#### 1. Login and Get Token

```bash
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "username": "admin",
    "password": "admin"
  }'
```

**Note:** Default credentials are `admin` / `admin`. In production, implement proper user authentication.

#### 2. Search Content

```bash
curl -X GET "http://localhost:8080/api/v1/search?query=golang&page=1&page_size=20" \
  -H "Authorization: Bearer YOUR_TOKEN_HERE"
```

#### 3. Search Video Content

```bash
curl -X GET "http://localhost:8080/api/v1/search?query=tutorial&content_type=video&sort_by=score" \
  -H "Authorization: Bearer YOUR_TOKEN_HERE"
```

#### 4. Get Content Details

```bash
curl -X GET "http://localhost:8080/api/v1/content/1" \
  -H "Authorization: Bearer YOUR_TOKEN_HERE"
```

#### 5. Health Check

```bash
curl -X GET http://localhost:8080/health
```

## Scoring Algorithm

The API sorts content by relevance score. The scoring algorithm considers the following factors:

### Score Components

1. **Base Score**

   - Video: `views / 1000 + likes / 100`
   - Text: `reading_time + reactions / 50`

2. **Type Coefficient**

   - Video: 1.5x
   - Text: 1.0x

3. **Freshness Score**

   - Within 1 week: +5
   - Within 1 month: +3
   - Within 3 months: +1
   - Older: +0

4. **Engagement Score**
   - Video: `(likes / views) × 10`
   - Text: `(reactions / reading_time) × 5`

### Final Score Formula

```
Final Score = (Base Score × Type Coefficient) + Freshness Score + Engagement Score
```

## OpenAPI/Swagger Documentation

For detailed API specification, you can access the OpenAPI documentation:

- **OpenAPI YAML**: `http://localhost:8080/docs/openapi.yaml` (requires authentication)
- **Swagger UI**: `http://localhost:8080/docs` (requires authentication)

You can test all endpoints and access interactive documentation through Swagger UI.

## Notes

- All date/time values are returned in ISO 8601 format in UTC timezone
- The `page` parameter for pagination starts from 1
- The `page_size` parameter must be between 1-100
- Search queries are case-insensitive
- Scores are calculated in real-time and cached
- Data from providers is automatically synchronized
