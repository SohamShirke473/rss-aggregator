# RSS Aggregator

A lightweight, concurrency-safe RSS feed aggregator backend built with Go, PostgreSQL, and Chi. It automatically scrapes RSS feeds on a background schedule and provides a REST API to manage feeds, follow sources, and retrieve posts.

## Quick Start

1. **Configure environment variables** in a `.env` file:
   ```env
   PORT=8080
   DATABASE_URL=postgres://user:password@localhost:5432/rssagg?sslmode=disable
   ```

2. **Run database migrations** (using [Goose](https://github.com/pressly/goose)):
   ```bash
   cd sql/schema
   goose postgres "$DATABASE_URL" up
   cd ../..
   ```

3. **Run the server**:
   ```bash
   go run .
   ```

The server will listen on `PORT` and run a background worker that fetches feeds every minute.

## API Endpoints

All endpoints use JSON. Authenticated endpoints require the header `Authorization: ApiKey <your_api_key>`.

| Method | Endpoint | Description | Auth Required |
|---|---|---|---|
| `GET` | `/v1/ready` | Health check endpoint | No |
| `GET` | `/v1/err` | Error test endpoint | No |
| `POST` | `/v1/users` | Register a new user and generate an API key | No |
| `GET` | `/v1/users` | Retrieve current authenticated user profile | Yes |
| `POST` | `/v1/feeds` | Add a new RSS feed to track | Yes |
| `GET` | `/v1/feeds` | List all tracked RSS feeds | No |
| `POST` | `/v1/feed_follows` | Follow a feed by its feed ID | Yes |
| `GET` | `/v1/feed_follows` | Get all feeds followed by current user | Yes |
| `DELETE` | `/v1/feed_follows/{feedFollowID}` | Unfollow a feed | Yes |
| `GET` | `/v1/posts` | Fetch latest scraped posts from followed feeds | Yes |
