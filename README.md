# RSS Aggregator Backend

A production-grade, concurrency-safe RSS feed aggregator API built with Go, PostgreSQL, and Chi. The service automatically scrapes RSS feeds on a background schedule and provides a RESTful API for users to manage feeds, follow sources, and retrieve aggregated post updates.

---

## Features

- **User Authentication**: Secure API key-based authentication (`ApiKey <token>` scheme).
- **Feed Management**: Add, track, and browse RSS feeds.
- **Feed Follows**: Follow and unfollow feeds dynamically per user account.
- **Concurrent RSS Scraping**: Background goroutine worker periodically fetches and parses RSS feeds concurrently.
- **Post Aggregation**: Automatically deduplicates and stores parsed articles and blog posts.
- **Type-Safe Database Access**: Built using **SQLC** with PostgreSQL for compile-time verified SQL queries.
- **Database Migrations**: Tracked and managed using **Goose**.
- **CORS Enabled**: Configured for seamless cross-origin requests.

---

## Tech Stack

- **Language**: [Go](https://go.dev/) (1.23+)
- **HTTP Router**: [go-chi/chi](https://github.com/go-chi/chi)
- **Database**: [PostgreSQL](https://www.postgresql.org/)
- **SQL Generator**: [sqlc](https://sqlc.dev/)
- **Database Migrations**: [Goose](https://github.com/pressly/goose)
- **Driver**: [lib/pq](https://github.com/lib/pq)

---

## Project Structure

```
rss-aggregator/
├── handler_err.go             # Error test handler
├── handler_feed.go            # Feed creation & listing handlers
├── handler_feed_follows.go    # Feed follow/unfollow handlers
├── handler_posts.go           # Post retrieval handler
├── handler_readiness.go       # Readiness check handler
├── handler_user.go            # User registration & profile handlers
├── json.go                    # JSON helper functions (marshal/error responses)
├── main.go                    # Server bootstrap, router setup & worker initialization
├── middleware_auth.go         # Authentication middleware
├── models.go                  # API response models & converters
├── scrape.go                  # Background RSS scraper & XML parser
├── sqlc.yaml                  # SQLC code generator configuration
├── internal/
│   ├── auth/                  # API key extraction logic
│   └── database/              # SQLC-generated database client & models
└── sql/
    ├── queries/               # SQL queries for SQLC
    └── schema/                # Goose migration SQL scripts
```

---

## Prerequisites

- **Go**: `1.23` or higher
- **PostgreSQL**: Running instance
- **Goose** (for running migrations):
  ```bash
  go install github.com/pressly/goose/v3/cmd/goose@latest
  ```
- **SQLC** (optional, for regenerating queries):
  ```bash
  go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest
  ```

---

## Environment Variables

Create a `.env` file in the project root with the following variables:

```env
PORT=8080
DATABASE_URL=postgres://<user>:<password>@localhost:5432/<dbname>?sslmode=disable
```

---

## Getting Started

### 1. Clone & Dependencies

```bash
git clone https://github.com/SohamShirke473/rss-aggregator.git
cd rss-aggregator
go mod download
```

### 2. Run Database Migrations

Apply the database migrations using Goose:

```bash
cd sql/schema
goose postgres "postgres://<user>:<password>@localhost:5432/<dbname>?sslmode=disable" up
cd ../..
```

### 3. Run the Server

```bash
go run .
```

The server will start on the configured `PORT` (e.g., `http://localhost:8080`) and the background scraper will automatically run every minute to fetch new feed posts.

---

## API Reference

All requests expecting or returning data use `application/json`. Authenticated routes require an `Authorization` header formatted as:

```http
Authorization: ApiKey <your_api_key>
```

### Health & System

#### 1. Readiness Check
- **Endpoint**: `GET /v1/ready`
- **Auth**: None
- **Response** (`200 OK`):
  ```json
  {}
  ```

#### 2. Error Check
- **Endpoint**: `GET /v1/err`
- **Auth**: None
- **Response** (`500 Internal Server Error`):
  ```json
  {
    "error": "Something went wrong"
  }
  ```

---

### Users

#### 1. Create User
- **Endpoint**: `POST /v1/users`
- **Auth**: None
- **Request Body**:
  ```json
  {
    "name": "Jane Doe"
  }
  ```
- **Response** (`201 Created`):
  ```json
  {
    "id": "e309cb69-7c18-472d-8ae5-0726d17df650",
    "created_at": "2026-10-19T03:15:00Z",
    "updated_at": "2026-10-19T03:15:00Z",
    "name": "Jane Doe",
    "api_key": "685c2c7f40880092c4bce8110b1062f689fbf..."
  }
  ```

#### 2. Get Current User Profile
- **Endpoint**: `GET /v1/users`
- **Auth**: Required (`Authorization: ApiKey <api_key>`)
- **Response** (`200 OK`):
  ```json
  {
    "id": "e309cb69-7c18-472d-8ae5-0726d17df650",
    "created_at": "2026-10-19T03:15:00Z",
    "updated_at": "2026-10-19T03:15:00Z",
    "name": "Jane Doe",
    "api_key": "685c2c7f40880092c4bce8110b1062f689fbf..."
  }
  ```

---

### Feeds

#### 1. Create Feed
- **Endpoint**: `POST /v1/feeds`
- **Auth**: Required (`Authorization: ApiKey <api_key>`)
- **Request Body**:
  ```json
  {
    "name": "TechCrunch",
    "url": "https://techcrunch.com/feed/"
  }
  ```
- **Response** (`201 Created`):
  ```json
  {
    "id": "a1b2c3d4-e5f6-7890-abcd-ef1234567890",
    "created_at": "2026-10-19T03:20:00Z",
    "updated_at": "2026-10-19T03:20:00Z",
    "name": "TechCrunch",
    "url": "https://techcrunch.com/feed/",
    "user_id": "e309cb69-7c18-472d-8ae5-0726d17df650"
  }
  ```

#### 2. Get All Feeds
- **Endpoint**: `GET /v1/feeds`
- **Auth**: None
- **Response** (`200 OK`):
  ```json
  [
    {
      "id": "a1b2c3d4-e5f6-7890-abcd-ef1234567890",
      "created_at": "2026-10-19T03:20:00Z",
      "updated_at": "2026-10-19T03:20:00Z",
      "name": "TechCrunch",
      "url": "https://techcrunch.com/feed/",
      "user_id": "e309cb69-7c18-472d-8ae5-0726d17df650"
    }
  ]
  ```

---

### Feed Follows

#### 1. Follow a Feed
- **Endpoint**: `POST /v1/feed_follows`
- **Auth**: Required (`Authorization: ApiKey <api_key>`)
- **Request Body**:
  ```json
  {
    "feed_id": "a1b2c3d4-e5f6-7890-abcd-ef1234567890"
  }
  ```
- **Response** (`201 Created`):
  ```json
  {
    "id": "f5e4d3c2-b1a0-9876-5432-10fedcba9876",
    "created_at": "2026-10-19T03:25:00Z",
    "updated_at": "2026-10-19T03:25:00Z",
    "user_id": "e309cb69-7c18-472d-8ae5-0726d17df650",
    "feed_id": "a1b2c3d4-e5f6-7890-abcd-ef1234567890"
  }
  ```

#### 2. Get Followed Feeds
- **Endpoint**: `GET /v1/feed_follows`
- **Auth**: Required (`Authorization: ApiKey <api_key>`)
- **Response** (`200 OK`):
  ```json
  [
    {
      "id": "f5e4d3c2-b1a0-9876-5432-10fedcba9876",
      "created_at": "2026-10-19T03:25:00Z",
      "updated_at": "2026-10-19T03:25:00Z",
      "user_id": "e309cb69-7c18-472d-8ae5-0726d17df650",
      "feed_id": "a1b2c3d4-e5f6-7890-abcd-ef1234567890"
    }
  ]
  ```

#### 3. Unfollow a Feed
- **Endpoint**: `DELETE /v1/feed_follows/{feedFollowID}`
- **Auth**: Required (`Authorization: ApiKey <api_key>`)
- **Response** (`200 OK`):
  ```json
  {}
  ```

---

### Posts

#### 1. Get Posts for User
Retrieves latest posts from feeds followed by the authenticated user.
- **Endpoint**: `GET /v1/posts`
- **Auth**: Required (`Authorization: ApiKey <api_key>`)
- **Response** (`200 OK`):
  ```json
  [
    {
      "id": "99e1a2b3-c4d5-6789-0123-456789abcdef",
      "created_at": "2026-10-19T03:30:00Z",
      "updated_at": "2026-10-19T03:30:00Z",
      "title": "Exciting Tech News",
      "description": "A summary of the latest updates in technology...",
      "published_at": "2026-10-19T03:00:00Z",
      "url": "https://techcrunch.com/2026/10/19/exciting-tech-news/",
      "feed_id": "a1b2c3d4-e5f6-7890-abcd-ef1234567890",
      "feed_name": "TechCrunch"
    }
  ]
  ```

---

## Database Migrations & SQL Generation

If you modify SQL queries in `sql/queries/` or migrations in `sql/schema/`:

- **Run migrations**:
  ```bash
  cd sql/schema
  goose postgres "<DATABASE_URL>" up
  ```
- **Roll back migrations**:
  ```bash
  cd sql/schema
  goose postgres "<DATABASE_URL>" down
  ```
- **Regenerate SQLC Go code**:
  ```bash
  sqlc generate
  ```
