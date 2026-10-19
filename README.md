# RSS Aggregator Backend

<p align="center">
  <img src="https://img.shields.io/badge/Go-1.23+-00ADD8?style=for-the-badge&logo=go&logoColor=white" alt="Go Version" />
  <img src="https://img.shields.io/badge/PostgreSQL-15+-4169E1?style=for-the-badge&logo=postgresql&logoColor=white" alt="PostgreSQL" />
  <img src="https://img.shields.io/badge/Router-Chi%20v5-blueviolet?style=for-the-badge" alt="Chi Router" />
  <img src="https://img.shields.io/badge/Database-SQLC-orange?style=for-the-badge" alt="SQLC" />
  <img src="https://img.shields.io/badge/Migrations-Goose-00C7B7?style=for-the-badge" alt="Goose" />
</p>

A production-grade, concurrency-safe RSS feed aggregator API built with **Go**, **PostgreSQL**, and **Chi**. The service automatically scrapes RSS feeds on a background schedule and provides a RESTful API for users to manage feeds, follow sources, and retrieve aggregated post updates.

---

## Table of Contents

- [Features](#features)
- [System Architecture](#system-architecture)
- [Database Entity-Relationship (ER) Model](#database-entity-relationship-er-model)
- [Tech Stack](#tech-stack)
- [Project Structure](#project-structure)
- [Prerequisites](#prerequisites)
- [Quick Start Guide](#quick-start-guide)
  - [1. Running PostgreSQL with Docker](#1-running-postgresql-with-docker-optional)
  - [2. Environment Configuration](#2-environment-configuration)
  - [3. Database Migrations](#3-database-migrations)
  - [4. Run the Server](#4-run-the-server)
- [End-to-End cURL Testing Walkthrough](#end-to-end-curl-testing-walkthrough)
- [API Reference](#api-reference)
  - [Authentication](#authentication)
  - [System Endpoints](#system-endpoints)
  - [User Endpoints](#user-endpoints)
  - [Feed Endpoints](#feed-endpoints)
  - [Feed Follow Endpoints](#feed-follow-endpoints)
  - [Post Endpoints](#post-endpoints)
- [Background Scraper Worker](#background-scraper-worker)
- [Database Migrations & SQL Generation](#database-migrations--sql-generation)
- [Contributing](#contributing)
- [License](#license)

---

## Features

- **User Authentication**: Secure API key-based authentication (`ApiKey <token>` header scheme) with SHA-256 generated keys.
- **Feed Management**: Register, track, and browse RSS feeds with unique URL constraints.
- **Feed Follows**: Dynamic follow/unfollow functionality per user account with cascading relationships.
- **Concurrent RSS Scraping**: Background goroutine worker periodically fetches and parses RSS feeds concurrently using `sync.WaitGroup`.
- **Post Aggregation & Deduplication**: Extracts post metadata (title, link, published timestamp, description) and skips duplicates using `ON CONFLICT (url) DO NOTHING`.
- **Type-Safe Database Access**: Built with **SQLC** for compile-time verified Go database queries.
- **Versioned Database Migrations**: Tracked and managed using **Goose**.
- **CORS Enabled**: Configured for seamless cross-origin requests across web frontends.

---

## System Architecture

```mermaid
flowchart TD
    Client[Client / Frontend] -->|HTTP Requests| Router[Chi Router /v1]
    
    subgraph API Gateway & Handlers
        Router --> AuthMiddleware[Auth Middleware]
        Router --> PublicHandlers[Public Endpoints: /ready, /feeds]
        AuthMiddleware --> ProtectedHandlers[Protected Handlers: /users, /feed_follows, /posts]
    end

    subgraph Business Logic & Storage
        PublicHandlers --> SQLC[SQLC Queries]
        ProtectedHandlers --> SQLC
        SQLC --> Postgres[(PostgreSQL Database)]
    end

    subgraph Background Scraper Worker
        Ticker[1-Minute Ticker] --> FetchFeeds[GetNextFeedsToFetch NULLS FIRST]
        FetchFeeds --> Postgres
        FetchFeeds --> FanOut[Goroutine Pool 10 Workers]
        FanOut --> HTTPGet[Fetch & Parse RSS XML]
        HTTPGet --> SavePosts[CreatePost ON CONFLICT DO NOTHING]
        SavePosts --> Postgres
        HTTPGet --> MarkFetched[Update last_fetched_at]
        MarkFetched --> Postgres
    end
```

---

## Database Entity-Relationship (ER) Model

```mermaid
erDiagram
    USERS ||--o{ FEEDS : "creates"
    USERS ||--o{ FEED_FOLLOWS : "follows"
    FEEDS ||--o{ FEED_FOLLOWS : "followed by"
    FEEDS ||--o{ POSTS : "contains"

    USERS {
        uuid id PK
        timestamp created_at
        timestamp updated_at
        text name
        varchar api_key UK
    }

    FEEDS {
        uuid id PK
        timestamp created_at
        timestamp updated_at
        text name
        text url UK
        uuid user_id FK
        timestamp last_fetched_at
    }

    FEED_FOLLOWS {
        uuid id PK
        timestamp created_at
        timestamp updated_at
        uuid user_id FK
        uuid feed_id FK
    }

    POSTS {
        uuid id PK
        timestamp created_at
        timestamp updated_at
        uuid feed_id FK
        text title
        text description
        timestamp published_at
        text url UK
    }
```

---

## Tech Stack

- **Language**: [Go](https://go.dev/) (1.23+)
- **HTTP Router**: [go-chi/chi](https://github.com/go-chi/chi) v5
- **Database**: [PostgreSQL](https://www.postgresql.org/)
- **SQL Generator**: [sqlc](https://sqlc.dev/)
- **Database Migrations**: [Goose](https://github.com/pressly/goose)
- **Driver**: [lib/pq](https://github.com/lib/pq)
- **Environment Management**: [godotenv](https://github.com/joho/godotenv)
- **UUID**: [google/uuid](https://github.com/google/uuid)

---

## Project Structure

```
rss-aggregator/
├── handler_err.go             # Error simulation handler
├── handler_feed.go            # Feed creation & listing handlers
├── handler_feed_follows.go    # Feed follow/unfollow handlers
├── handler_posts.go           # Post retrieval handler
├── handler_readiness.go       # Readiness check handler
├── handler_user.go            # User registration & profile handlers
├── json.go                    # JSON helper functions (marshal & error payloads)
├── main.go                    # Server bootstrap, router setup & worker initialization
├── middleware_auth.go         # Authentication middleware
├── models.go                  # API response models & type converters
├── scrape.go                  # Background RSS scraper & XML parser
├── sqlc.yaml                  # SQLC code generator configuration
├── internal/
│   ├── auth/                  # API key extraction logic from headers
│   └── database/              # SQLC-generated database client & models
└── sql/
    ├── queries/               # SQL queries for SQLC
    │   ├── feed_follows.sql
    │   ├── feeds.sql
    │   ├── posts.sql
    │   └── users.sql
    └── schema/                # Goose migration SQL scripts
        ├── 001_users.sql
        ├── 002_users.sql
        ├── 003_feeds.sql
        ├── 004_feed_follows.sql
        ├── 005_feeds.sql
        └── 006_posts.sql
```

---

## Prerequisites

- **Go**: `1.23` or higher installed
- **PostgreSQL**: Running instance (local or via Docker)
- **Goose** (for running migrations):
  ```bash
  go install github.com/pressly/goose/v3/cmd/goose@latest
  ```
- **SQLC** (optional, for regenerating queries):
  ```bash
  go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest
  ```

---

## Quick Start Guide

### 1. Running PostgreSQL with Docker (Optional)

If you don't have a local PostgreSQL instance running:

```bash
docker run --name rssagg-postgres \
  -e POSTGRES_USER=postgres \
  -e POSTGRES_PASSWORD=postgres \
  -e POSTGRES_DB=rssagg \
  -p 5432:5432 -d postgres:15-alpine
```

### 2. Environment Configuration

Create a `.env` file in the project root:

```env
PORT=8080
DATABASE_URL=postgres://postgres:postgres@localhost:5432/rssagg?sslmode=disable
```

### 3. Database Migrations

Apply the database migrations using Goose:

```bash
cd sql/schema
goose postgres "postgres://postgres:postgres@localhost:5432/rssagg?sslmode=disable" up
cd ../..
```

### 4. Run the Server

```bash
go run .
```

The server will start on port `8080` (or the configured `PORT`), and the background worker will begin scraping feeds every minute.

---

## End-to-End cURL Testing Walkthrough

Follow this step-by-step walkthrough in your terminal to test the entire lifecycle:

```bash
# 1. Health check
curl http://localhost:8080/v1/ready

# 2. Create a user (Copy the 'api_key' from response)
curl -X POST http://localhost:8080/v1/users \
  -H "Content-Type: application/json" \
  -d '{"name": "Alice"}'

# Export your API key for subsequent requests
API_KEY="<paste_api_key_here>"

# 3. Retrieve your user profile
curl http://localhost:8080/v1/users \
  -H "Authorization: ApiKey $API_KEY"

# 4. Add an RSS feed (e.g. Wagslane Blog)
curl -X POST http://localhost:8080/v1/feeds \
  -H "Authorization: ApiKey $API_KEY" \
  -H "Content-Type: application/json" \
  -d '{"name": "Wagslane", "url": "https://wagslane.dev/index.xml"}'

# 5. List all registered feeds
curl http://localhost:8080/v1/feeds

# 6. Follow the feed (Copy the 'id' of the feed from step 5)
curl -X POST http://localhost:8080/v1/feed_follows \
  -H "Authorization: ApiKey $API_KEY" \
  -H "Content-Type: application/json" \
  -d '{"feed_id": "<paste_feed_id_here>"}'

# 7. Check your followed feeds
curl http://localhost:8080/v1/feed_follows \
  -H "Authorization: ApiKey $API_KEY"

# 8. Fetch aggregated posts (Scraper collects posts within 1 minute)
curl http://localhost:8080/v1/posts \
  -H "Authorization: ApiKey $API_KEY"
```

---

## API Reference

### Authentication

Protected endpoints require the `Authorization` header formatted as:

```http
Authorization: ApiKey <your_api_key>
```

| HTTP Status | Meaning |
|---|---|
| `200 OK` | Request succeeded |
| `201 Created` | Resource created successfully |
| `400 Bad Request` | Missing/malformed JSON or query parameters |
| `403 Forbidden` | Missing or invalid API key |
| `500 Internal Error` | Server error |

---

### System Endpoints

#### Readiness Check
- **Route**: `GET /v1/ready`
- **Auth**: None
- **Response** (`200 OK`):
  ```json
  {}
  ```

#### Error Simulation Check
- **Route**: `GET /v1/err`
- **Auth**: None
- **Response** (`500 Internal Server Error`):
  ```json
  {
    "error": "Something went wrong"
  }
  ```

---

### User Endpoints

#### Create User
- **Route**: `POST /v1/users`
- **Auth**: None
- **Body**:
  ```json
  {
    "name": "Alice"
  }
  ```
- **Response** (`201 Created`):
  ```json
  {
    "id": "e309cb69-7c18-472d-8ae5-0726d17df650",
    "created_at": "2026-10-19T03:15:00Z",
    "updated_at": "2026-10-19T03:15:00Z",
    "name": "Alice",
    "api_key": "685c2c7f40880092c4bce8110b1062f689fbf..."
  }
  ```

#### Get Current User Profile
- **Route**: `GET /v1/users`
- **Auth**: Required (`Authorization: ApiKey <api_key>`)
- **Response** (`200 OK`):
  ```json
  {
    "id": "e309cb69-7c18-472d-8ae5-0726d17df650",
    "created_at": "2026-10-19T03:15:00Z",
    "updated_at": "2026-10-19T03:15:00Z",
    "name": "Alice",
    "api_key": "685c2c7f40880092c4bce8110b1062f689fbf..."
  }
  ```

---

### Feed Endpoints

#### Create Feed
- **Route**: `POST /v1/feeds`
- **Auth**: Required (`Authorization: ApiKey <api_key>`)
- **Body**:
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

#### Get All Feeds
- **Route**: `GET /v1/feeds`
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

### Feed Follow Endpoints

#### Follow a Feed
- **Route**: `POST /v1/feed_follows`
- **Auth**: Required (`Authorization: ApiKey <api_key>`)
- **Body**:
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

#### Get Followed Feeds
- **Route**: `GET /v1/feed_follows`
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

#### Unfollow a Feed
- **Route**: `DELETE /v1/feed_follows/{feedFollowID}`
- **Auth**: Required (`Authorization: ApiKey <api_key>`)
- **Response** (`200 OK`):
  ```json
  {}
  ```

---

### Post Endpoints

#### Get Posts for User
Retrieves up to 10 latest posts from feeds followed by the authenticated user, sorted in descending order of publish date.
- **Route**: `GET /v1/posts`
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

## Background Scraper Worker

The background worker in `scrape.go` runs continuously on an interval:

1. **Selection**: Queries `GetNextFeedsToFetch` ordering by `last_fetched_at ASC NULLS FIRST` with a batch limit (default: 10).
2. **Concurrency**: Spawns concurrent goroutines monitored with a `sync.WaitGroup`.
3. **HTTP Fetch**: Issues a `GET` request with a 10-second timeout.
4. **XML Parsing**: Unmarshals the RSS XML feed structure (`<channel>`, `<item>`).
5. **Deduplication**: Inserts parsed posts with `ON CONFLICT (url) DO NOTHING` to ensure idempotency.
6. **Timestamp Update**: Marks the feed as fetched by updating `last_fetched_at`.

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
- **Check migration status**:
  ```bash
  cd sql/schema
  goose postgres "<DATABASE_URL>" status
  ```
- **Regenerate SQLC Go code**:
  ```bash
  sqlc generate
  ```

---

