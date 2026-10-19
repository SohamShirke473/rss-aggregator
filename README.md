# RSS Aggregator

A lightweight RSS feed aggregator backend built with Go and PostgreSQL.

## Quick Start

1. Configure environment variables in `.env`:
   ```env
   PORT=8080
   DATABASE_URL=postgres://user:password@localhost:5432/rssagg?sslmode=disable
   ```

2. Run the server:
   ```bash
   go run .
   ```
