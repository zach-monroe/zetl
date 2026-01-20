# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

Zetl is a quote management web application with a Go backend (Gin framework) and HTML/HTMX frontend. Users can store, browse, and manage quotes with metadata (author, book, tags, notes). Features session-based authentication, user profiles, and interactive card UI with flip animations.

## Development Commands

All commands run from `server/` directory:

```bash
make dev              # Hot-reload for Go (air) + TailwindCSS watch mode
make air              # Go server with hot-reload only
make tailwind         # TailwindCSS watch mode only
make tailwind-build   # One-time minified CSS build
```

One-time CSS build (no watch):
```bash
npx @tailwindcss/cli -i ../client/src/input.css -o ../client/css/style.css
```

Server runs on `localhost:8080`

## Architecture

### Backend (`server/`)

```
main.go              # Router setup, route definitions, session config
handlers/            # HTTP handlers by domain
  auth_handler.go    # Login, signup, logout, password reset
  quote_handler.go   # CRUD for quotes
  page_handlers.go   # HTML page rendering
  settings_handler.go
middleware/
  auth.go            # AuthRequired(), QuoteOwnershipRequired()
database/
  database.go        # DB connection, legacy quote queries
  user_queries.go    # User CRUD
  quote_queries.go   # Quote CRUD
models/
  models.go          # Quote struct
  user.go            # User, PrivacySettings structs
services/
  auth_service.go    # Password hashing (bcrypt)
  email_service.go   # Password reset emails
```

### Frontend (`client/`)

```
templates/           # Go HTML templates
  base.html          # Header, navigation, filter dropdown (shared partial)
  index.html         # Home page with quote grid, modals
  profile.html       # User profile with their quotes
  login.html, signup.html, settings.html, etc.
js/
  main.js            # All client-side JS (card interactions, modals, search, filtering)
src/
  input.css          # TailwindCSS source
css/
  style.css          # Compiled CSS (don't edit directly)
```

### Key Patterns

**Template inheritance**: Templates use Go's `{{ define }}` and `{{ template }}` for partials. `base.html` defines `header` and `header-scripts` blocks included in other pages.

**Session authentication**: Uses `gin-contrib/sessions` with cookie store. User ID stored in session, retrieved via `middleware.AuthRequired()`.

**Quote ownership**: `middleware.QuoteOwnershipRequired()` checks if logged-in user owns the quote before allowing edit/delete.

**Client-side interactivity**: `main.js` handles card flip animations, hover expansion (FLIP technique), 3-dot menus, modals, fuzzy tag search, and filtering. Functions exposed globally via `window.functionName` for onclick handlers.

**Dynamic user context**: Body tag has `data-user-id` attribute set by Go template, read by JS for ownership checks.

## API Routes

**Public**: `GET /`, `GET /login`, `GET /signup`, `GET /user/:id/quotes`

**Auth** (`/auth`): `POST /login`, `POST /signup`, `POST /logout`, `POST /forgot-password`, `POST /reset-password`

**Protected** (`/api`, requires auth):
- `GET /user` - current user
- `PUT /user/profile`, `PUT /user/password`, `PUT /user/privacy`
- `POST /quote` - create quote
- `PUT /quote/:id`, `DELETE /quote/:id` - requires ownership

## Database

PostgreSQL with environment variables from `server/.env`:
- `DB_HOSTNAME`, `DB_USERNAME`, `DB_PASSWORD`, `DB_PORT`, `DB_NAME`

Key tables: `users`, `quotes`, `password_reset_tokens`

Tags stored as PostgreSQL `text[]` arrays, converted to/from Go `[]string` in database layer.
