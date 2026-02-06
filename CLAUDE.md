# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

Zetl is a quote management web application with a Go backend (Gin framework) and HTML/HTMX frontend. Users can store, browse, manage, and share quotes with metadata (author, book, tags, notes). Features session-based authentication, user profiles with privacy controls, password reset via email, and an interactive card UI with flip animations and FLIP-based hover expansion.

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
main.go              # Router setup, route definitions, PostgreSQL session store
handlers/
  auth_handler.go    # Login, signup, logout (auto-login after signup)
  quote_handler.go   # CRUD for quotes
  page_handlers.go   # HTML page rendering with template context
  settings_handler.go # Profile, password, privacy updates
  password_reset_handler.go  # Forgot/reset password flow
middleware/
  auth.go            # AuthRequired(), OptionalAuth(), QuoteOwnershipRequired()
database/
  database.go        # DB connection, tag array conversion
  user_queries.go    # User CRUD, last login tracking
  quote_queries.go   # Quote CRUD, ownership verification
  password_reset_queries.go  # Token generation, validation, cleanup
models/
  models.go          # Quote struct
  user.go            # User, PrivacySettings, request/response types
services/
  auth_service.go    # Password hashing (bcrypt cost 12), validation
  email_service.go   # SMTP with STARTTLS (Gmail compatible)
```

### Frontend (`client/`)

```
templates/
  base.html          # Header, navigation, filter dropdown, quote-cards partial
  index.html         # Home page with quote grid, add/edit/delete modals
  profile.html       # User profile with bio, quotes, edit link
  settings.html      # Profile/password/privacy sections with per-section feedback
  login.html         # Username or email + password
  signup.html        # Registration with password confirmation
  forgot-password.html
  reset-password.html
js/
  main.js            # Card flip, FLIP expansion, menus, modals, fuzzy search, filtering
src/
  input.css          # TailwindCSS source
css/
  style.css          # Compiled CSS (don't edit directly)
```

### Key Patterns

**Session authentication**: Uses `gin-contrib/sessions` with PostgreSQL-backed store. 24-hour expiration, HttpOnly cookies, SameSite=Lax. Session secret from `SESSION_SECRET` env var.

**Middleware chain**:
- `AuthRequired()` - Verifies session, sets `user_id` in context, returns 401 if missing
- `OptionalAuth()` - Sets `user_id` if logged in but doesn't require it
- `QuoteOwnershipRequired(db)` - Verifies user owns the quote, returns 403 if not

**Template inheritance**: Templates use `{{ define }}` and `{{ template }}` for partials. `base.html` defines `header`, `header-scripts`, and `quote-cards` blocks.

**Card animations**: `main.js` implements FLIP technique (First-Last-Invert-Play) for smooth card repositioning during hover expansion. 500ms animation duration, bounce easing for expand, smooth for contract.

**Tag filtering**: Dropdown-only design with fuzzy matching algorithm. Exact match scores 1000, starts-with 500+, contains 200+, fuzzy chars 10+ with consecutive bonus. Applies AND logic (card must have ALL selected tags).

**Dynamic user context**: Body tag has `data-user-id` attribute set by Go template, read by JS for ownership checks on card menus.

## API Routes

**Public**: `GET /`, `GET /login`, `GET /signup`, `GET /forgot-password`, `GET /reset-password`, `GET /user/:id/quotes`

**Auth** (`/auth`):
- `POST /signup` - Creates user, auto-logs in
- `POST /login` - Accepts username OR email, updates last_login
- `POST /logout` - Clears session
- `POST /forgot-password` - Sends reset email (always returns success to prevent enumeration)
- `POST /reset-password` - Validates token, updates password, auto-logs in

**Protected Pages** (requires auth): `GET /settings`, `GET /profile`

**Protected API** (`/api`, requires auth):
- `GET /user` - Current user info
- `PUT /user/profile` - Update username, email, bio
- `PUT /user/password` - Requires current password verification
- `PUT /user/privacy` - Toggle profile_public, quotes_public
- `POST /quote` - Create quote
- `PUT /quote/:id`, `DELETE /quote/:id` - Requires ownership

## Database

PostgreSQL with environment variables from `server/.env`:
- `DB_HOSTNAME`, `DB_USERNAME`, `DB_PASSWORD`, `DB_PORT`, `DB_NAME`

Key tables: `users`, `quotes`, `password_reset_tokens`

**Tags**: Stored as PostgreSQL `text[]` arrays (e.g., `{tag1,tag2}`), converted to/from Go `[]string` via custom parsing. Uses `pq.Array()` for inserts.

**Privacy settings**: Stored as JSON column in users table with `profile_public` and `quotes_public` booleans.

**Password reset tokens**: 64-char hex tokens, 1-hour expiration, single-use. Old tokens invalidated when new one created or password changed.

## Email Configuration

SMTP settings in `server/.env`:
- `SMTP_HOST`, `SMTP_PORT`, `SMTP_USERNAME`, `SMTP_PASSWORD`, `SMTP_FROM`, `APP_URL`

Uses STARTTLS for Gmail compatibility (port 587). Falls back to logging reset links if not configured.

## Models

**Quote**: QuoteID, UserID, Quote, Author, Book, Tags ([]string), Notes

**User**: ID, Username, Email, PasswordHash, Bio, PrivacySettings, CreatedAt, UpdatedAt, LastLogin, IsActive

**Password validation**: Min 8 chars, requires uppercase, lowercase, and digit.

## Infrastructure

### Kubernetes Deployment

The application runs on a 2-node K3s cluster:
- **Control plane node**: Runs workloads and cluster management
- **Worker node**: Runs workloads and bare-metal PostgreSQL

Deployment files in `k8s/`:
```
deployment.yaml      # Zetl application deployment
service.yaml         # ClusterIP service
ingress.yaml         # Traefik ingress with TLS
secret.yaml          # Environment variables (gitignored)
```

**Ingress**: Traefik with automatic TLS via cert-manager (letsencrypt-prod ClusterIssuer).

**Storage**: Uses K3s `local-path` provisioner for PersistentVolumes.

### Monitoring Stack

Deployed via Helm (`prometheus-community/kube-prometheus-stack`) in `monitoring/` namespace:

```
monitoring/
  values.yaml                    # Helm values for the stack
  postgres-exporter/
    deployment.yaml              # Scrapes PostgreSQL metrics
    service.yaml                 # ClusterIP on port 9187
    service-monitor.yaml         # Prometheus auto-discovery
    secret.yaml                  # DB connection string (gitignored)
```

Components:
- **Prometheus**: 15-day retention, 20Gi storage
- **Grafana**: Exposed via ingress with TLS
- **node-exporter**: DaemonSet on all nodes
- **postgres-exporter**: Custom deployment for database metrics

### Deployment Commands

```bash
# Deploy application
kubectl apply -f k8s/

# Install/upgrade monitoring stack
helm upgrade --install kube-prometheus-stack prometheus-community/kube-prometheus-stack \
  --namespace monitoring \
  --values monitoring/values.yaml

# Deploy postgres-exporter
kubectl apply -f monitoring/postgres-exporter/

# Check pod status
kubectl get pods -n default
kubectl get pods -n monitoring

# Check certificates
kubectl get certificate -A
```

### Hairpin NAT

LAN devices cannot access the application via public IP due to hairpin NAT limitations. CoreDNS is configured with NodeHosts to resolve domains directly to the Traefik ingress ClusterIP for in-cluster and LAN traffic.

## Security Notes

- Never commit `.env` or `secret.yaml` files (both are gitignored)
- Use `server/.env.example` as a template for local development
- Kubernetes secrets should be created manually on the cluster, not stored in git
- Password hashing uses bcrypt with cost factor 12
- Session cookies are HttpOnly with SameSite=Lax
