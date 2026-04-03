# Zetl

A quote management web application for storing, organizing, and sharing your favorite quotes. Built with Go, PostgreSQL, and HTMX, deployed on a self-hosted Kubernetes cluster with a full GitOps pipeline.

**Live Demo:** [zetl.zachmonroe.cloud](https://zetl.zachmonroe.cloud)

---

## Overview

Zetl is a full-stack web application that allows users to collect and manage quotes with rich metadata including author, book, tags, and personal notes. The application features an interactive card-based UI with flip animations, fuzzy search, tag filtering, and user profiles with privacy controls.

This project demonstrates end-to-end software development — from local development through production Kubernetes deployment — including a GitOps delivery pipeline, infrastructure monitoring, TLS certificate automation, and secure authentication patterns. A Raspberry Pi client uses Claude AI to OCR handwritten notecards and post them directly to the API.

---

## Tech Stack

| Layer | Technology |
|-------|------------|
| **Backend** | Go 1.21+, Gin web framework |
| **Database** | PostgreSQL 15 with array types for tags |
| **Frontend** | HTML templates, HTMX, TailwindCSS |
| **Authentication** | Session-based with PostgreSQL store, bcrypt hashing |
| **Infrastructure** | K3s (lightweight Kubernetes), 2-node cluster |
| **Ingress** | Traefik with automatic TLS via cert-manager |
| **CI/CD** | GitHub Actions → Docker Hub → ArgoCD (GitOps) |
| **Monitoring** | Prometheus, Grafana, node-exporter, postgres-exporter |
| **Pi Client** | Python, Claude Haiku (vision), fswebcam |
| **Email** | SMTP with STARTTLS (Gmail compatible) |

---

## Architecture

```
┌──────────┐  push   ┌──────────────────┐  update tag   ┌──────────────────┐
│  GitHub  │────────▶│  GitHub Actions  │──────────────▶│  kustomization   │
└──────────┘         │  (build & push)  │               │  (k8s/  in git)  │
                     └──────────────────┘               └────────┬─────────┘
                              │                                   │ sync
                              ▼                                   ▼
                     ┌──────────────────┐               ┌──────────────────┐
                     │    Docker Hub    │               │     ArgoCD       │
                     │  zachmonroe/zetl │               │  (auto-deploy)   │
                     └──────────────────┘               └──────────────────┘

┌──────────┐  HTTPS  ┌──────────────────┐  ┌─────────────────┐  ┌─────────────┐
│  Client  │────────▶│ Traefik Ingress  │─▶│   Zetl Pod      │─▶│  PostgreSQL │
│(Browser) │         │ (TLS via ACME)   │  │   (Go/Gin)      │  │ (Bare metal)│
└──────────┘         └──────────────────┘  └─────────────────┘  └─────────────┘

┌──────────────┐  webcam  ┌─────────────────────────────────────────┐
│  Raspberry   │─────────▶│  capture.py                             │
│  Pi Client   │          │  1. fswebcam → JPEG                     │
└──────────────┘          │  2. Claude Haiku (vision OCR) → JSON    │
                          │  3. POST /api/device/quote              │
                          └─────────────────────────────────────────┘

Cluster: 2-node K3s (control plane + worker)
```

---

## Features

- **Quote Management**: Create, edit, and delete quotes with author, book, tags, and notes
- **Interactive Card UI**: Flip animations reveal quote details; FLIP-based hover expansion for smooth repositioning
- **Tag System**: PostgreSQL array storage, fuzzy search filtering, AND logic for multi-tag queries
- **User Profiles**: Customizable bio, privacy controls (public/private profile and quotes)
- **Authentication**: Session-based auth with secure cookies, password reset via email
- **Raspberry Pi Client**: Capture a photo of a handwritten notecard; Claude AI extracts the quote fields and posts them to the API automatically
- **GitOps Deployment**: GitHub Actions builds and pushes a Docker image on every merge; ArgoCD detects the new image tag and deploys it to the cluster
- **Responsive Design**: Mobile-friendly TailwindCSS styling

---

## Raspberry Pi Client

The `pi-client/` directory contains a Python script that turns a Raspberry Pi with a webcam into a quote-capture station:

1. **Capture** — `fswebcam` takes a 1080p photo of a handwritten notecard
2. **OCR** — The image is sent to Claude Haiku (vision) with a structured prompt; it returns JSON with `quote`, `author`, `book`, `tags`, and `notes`
3. **Post** — The extracted data is posted to `/api/device/quote` using a bearer token

```bash
python capture.py         # Capture once and exit
python capture.py --loop  # Press Enter between captures (batch mode)
```

Requires a `.env` with `ZETL_URL`, `API_TOKEN`, and `ANTHROPIC_API_KEY`.

---

## CI/CD Pipeline

Every push to `main` triggers a GitHub Actions workflow:

1. Builds a Docker image tagged with the short commit SHA
2. Pushes it to Docker Hub (`zachmonroe/zetl:<sha>`)
3. Updates `k8s/kustomization.yaml` with the new tag and commits it back (`[skip ci]`)
4. ArgoCD detects the change and automatically syncs the cluster — zero manual deployment steps

---

## Security

| Feature | Implementation |
|---------|----------------|
| **Password Hashing** | bcrypt with cost factor 12 |
| **Session Management** | HttpOnly cookies, SameSite=Lax, 24-hour expiration |
| **Session Storage** | PostgreSQL-backed (not client-side) |
| **Password Reset** | 64-character hex tokens, 1-hour expiration, single-use |
| **Enumeration Prevention** | Password reset always returns success regardless of email existence |
| **Input Validation** | Server-side validation for all user inputs |
| **Ownership Verification** | Middleware checks quote ownership before edit/delete |
| **TLS** | Automatic certificate provisioning via Let's Encrypt |
| **Secrets Management** | Environment variables, gitignored credential files |

---

## Technical Challenges Solved

### Hairpin NAT Resolution

**Problem**: Devices on the local network couldn't access the application via its public domain because traffic would route through the router and fail to return properly.

**Solution**: Configured K3s CoreDNS with NodeHosts to resolve the domain directly to the Traefik ingress ClusterIP. This bypasses the public IP entirely for in-cluster and LAN requests.

### Kubernetes Deployment Strategy

**Challenge**: Deploying a stateful application with an external database on a minimal 2-node K3s cluster.

**Approach**:
- K3s `local-path` provisioner for persistent volumes
- PostgreSQL runs bare-metal on the worker node; application connects via static LAN IP
- Traefik handles ingress with automatic TLS via cert-manager's HTTP-01 challenge
- Helm-managed monitoring stack (kube-prometheus-stack) with custom values for K3s compatibility
- Disabled unreachable scrape targets (`kubeControllerManager`, `kubeScheduler`, `kubeEtcd`, `kubeProxy`) since K3s bundles these into a single binary

### Session-Based Authentication

**Challenge**: Implementing secure, scalable authentication without JWT complexity.

**Solution**:
- PostgreSQL-backed session store via `gin-contrib/sessions`
- Sessions survive server restarts and scale across replicas
- Middleware chain: `AuthRequired()` → `QuoteOwnershipRequired(db)` for protected routes
- Auto-login after signup and password reset for seamless UX
- Username OR email login flexibility

### Kernel Module Compatibility (Arch Linux)

**Problem**: After an Arch Linux kernel update, kube-proxy failed with "Extension statistic revision 0 not supported, missing kernel module?" because the running kernel didn't match installed modules.

**Solution**: Identified the mismatch by comparing `uname -r` with installed kernel packages, then rebooted to load the correct modules — a common gotcha with rolling-release distributions in production.

---

## Monitoring Stack

Deployed via Helm (`prometheus-community/kube-prometheus-stack`):

- **Prometheus**: 15-day retention, 20Gi storage, scrapes all ServiceMonitors
- **Grafana**: Available at [grafana.zachmonroe.cloud](https://grafana.zachmonroe.cloud) with custom dashboards for PostgreSQL and ArgoCD metrics
- **node-exporter**: DaemonSet collecting host metrics from both nodes
- **postgres-exporter**: Custom deployment scraping PostgreSQL metrics
- **Alertmanager**: Configured for future alerting rules

---

## AI-Assisted Development

This project was developed collaboratively with Claude Code (Anthropic's CLI tool):

- **Human-Written Foundation**: Core API structure, database schema, and frontend templates were written manually to establish architectural direction.
- **Iterative Enhancement**: Claude Code extended functionality with specific, detailed prompts — e.g., "Add password reset with email verification, 1-hour token expiration, and single-use tokens."
- **Infrastructure Automation**: A custom `homelab-health-monitor` agent handles cluster SSH, kubectl commands, health checks, and deployment verification, reducing context-switching during infrastructure work.

### Lessons Learned

- Specific prompts yield better results: "Add a logout button to the header nav, POST to `/auth/logout`, clear the session, redirect to home" outperforms "Add a logout button."
- `CLAUDE.md` with project conventions prevented style drift and reduced back-and-forth corrections.
- AI excels at boilerplate (Kubernetes manifests, CRUD handlers, repetitive patterns); humans excel at architecture (schema design, security decisions, UX flow).

---

## Local Development

```bash
# Clone the repository
git clone https://github.com/zach-monroe/zetl.git
cd zetl

# Set up environment variables
cp server/.env.example server/.env
# Edit .env with your PostgreSQL and SMTP credentials

# Install dependencies and run
cd server
make dev  # Starts Go server with hot-reload + TailwindCSS watch
```

Server runs at `http://localhost:8080`

### Prerequisites

- Go 1.21+
- PostgreSQL 15+
- Node.js (for TailwindCSS CLI)

---

## Try It Out

Visit [zetl.zachmonroe.cloud](https://zetl.zachmonroe.cloud):

1. Create an account (no email verification required for signup)
2. Add your favorite quotes with author, book, and tags
3. Click a card to flip and reveal details
4. Hover over cards to see the expansion animation
5. Use the tag filter to find specific quotes
6. Customize your profile and privacy settings

Feedback and suggestions are welcome!

---

## License

MIT License — see LICENSE file for details.

---

Created by [Zach Monroe](https://github.com/zach-monroe)
