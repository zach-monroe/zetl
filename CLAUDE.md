# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

Zetl is a quote management web application with a Go backend (Gin framework) and HTML/HTMX frontend. The application stores quotes with metadata (author, book, tags) in a PostgreSQL database and displays them in a card-based layout with TailwindCSS styling.

## Development Commands

### Running the Application

**Development mode** (hot-reload for both Go and Tailwind):
```bash
cd server
make dev
```

This runs both:
- `air` for Go hot-reloading (watches `.go`, `.html`, `.tpl`, `.tmpl` files)
- TailwindCSS in watch mode for CSS compilation

**Individual services**:
```bash
cd server
make air        # Run Go server with hot-reload only
make tailwind   # Run TailwindCSS watch mode only
```

**Standard build and run**:
```bash
cd server
go build -o tmp/main .
./tmp/main
```

Server runs on `localhost:8080`

### Dependencies

Install Go dependencies:
```bash
cd server
go mod download
```

## Architecture

### Backend Structure (`server/`)

- **main.go**: Application entry point
  - `setupRouter()`: Configures Gin routes and middleware
  - Root endpoint `/` serves HTML with quotes from database
  - `POST /quote`: Creates new quotes
  - `/admin` endpoint group with basic auth (credentials in code: foo/bar, manu/123)

- **database/database.go**: PostgreSQL connection and queries
  - `StartDatabase()`: Loads `.env` from `server/.env` and establishes connection
  - `FetchQuotesAsJson()`: Queries all quotes and converts to JSON (handles PostgreSQL array types for tags)
  - `AddQuoteToDatabase()`: Inserts new quotes into database

- **models/models.go**: Data structures
  - `Quote`: Represents a single quote with fields: quote_id, user_id, quote, author, book, tags
  - `Quotes`: Array of Quote

### Frontend Structure (`client/`)

- **index.html**: Go template with three defined blocks:
  - `index`: Main HTML structure, includes HTMX library
  - `cards`: Container and grid layout
  - `items`: Iterates through quotes and renders cards

- **TailwindCSS**: Input file at `client/src/input.css`, compiled to `client/css/style.css`

### Database Configuration

PostgreSQL connection uses environment variables from `server/.env`:
- `DB_HOSTNAME`: Database host
- `DB_USERNAME`: Database user
- `DB_PASSWORD`: Database password
- `DB_PORT`: Database port (typically 5432)
- `DB_NAME`: Database name

Database schema (quotes table):
- `quote_id` (int, primary key)
- `user_id` (int)
- `quote` (text)
- `author` (text)
- `book` (text)
- `tags` (text[], PostgreSQL array)

### Key Implementation Details

1. **Tags handling**: PostgreSQL stores tags as text arrays (`text[]`). The database layer converts between PostgreSQL array format (`{tag1,tag2}`) and Go string slices.

2. **Quote creation flow**:
   - POST request to `/quote` endpoint
   - Raw JSON body read directly
   - Passed to `AddQuoteToDatabase()` which parses and inserts into PostgreSQL

3. **Template rendering**: Server fetches all quotes on startup, unmarshals to Go structs, and passes to Gin HTML template for rendering.

4. **Static files**: CSS served from `../client/css` relative to server directory.

5. **Hot reload**: Air config (`.air.toml`) watches for changes and rebuilds binary to `tmp/main`, excludes test files and `tmp/` directory.
