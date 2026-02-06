# Zetl

A quote management web application for storing, organizing, and sharing your favorite quotes. Built with Go, PostgreSQL, and HTMX, deployed on a self-hosted Kubernetes cluster.

**Live Demo:** [zetl.zachmonroe.cloud](https://zetl.zachmonroe.cloud)

---

## Overview

Zetl is a full-stack web application that allows users to collect and manage quotes with rich metadata including author, book, tags, and personal notes. The application features an interactive card-based UI with flip animations, fuzzy search, tag filtering, and user profiles with privacy controls.

This project demonstrates end-to-end software development from local development through production Kubernetes deployment, including infrastructure monitoring, TLS certificate automation, and secure authentication patterns.

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
| **Monitoring** | Prometheus, Grafana, node-exporter, postgres-exporter |
| **Email** | SMTP with STARTTLS (Gmail compatible) |

---

## Architecture

```
┌──────────┐    HTTPS    ┌──────────────────────┐    ┌─────────────────┐    ┌─────────────┐
│  Client  │────────────▶│  Traefik Ingress     │───▶│   Zetl Pod      │───▶│  PostgreSQL │
│ (Browser)│             │  (TLS termination)   │    │   (Go/Gin)      │    │ (Bare metal)│
└──────────┘             └──────────────────────┘    └─────────────────┘    └─────────────┘
                                    │                        │
                         ┌──────────┴───────────┐            │
                         │  cert-manager        │            ▼
                         │  (Let's Encrypt)     │    ┌─────────────────┐
                         └──────────────────────┘    │  Session Store  │
                                    │                │  (PostgreSQL)   │
                         ┌──────────┴───────────┐    └─────────────────┘
                         │  Prometheus Stack    │
                         │  ├─ Prometheus       │◀──── Scrapes metrics from all components
                         │  ├─ Grafana          │
                         │  ├─ node-exporter    │
                         │  └─ postgres-exporter│
                         └──────────────────────┘

Cluster: 2-node K3s (control plane + worker)
```

---

## Features

### Current

- **Quote Management**: Create, edit, and delete quotes with author, book, tags, and notes
- **Interactive Card UI**: Flip animations reveal quote details; FLIP-based hover expansion for smooth repositioning
- **Tag System**: PostgreSQL array storage, fuzzy search filtering, AND logic for multi-tag queries
- **User Profiles**: Customizable bio, privacy controls (public/private profile and quotes)
- **Authentication**: Session-based auth with secure cookies, password reset via email
- **Responsive Design**: Mobile-friendly TailwindCSS styling

### Planned

- **Handwritten Quote Scanner**: Local LLM to OCR handwritten quotes and automatically populate the database via API
- **Custom Model Training**: PyTorch-based fine-tuning for improved handwriting recognition, eventually self-hosted on AWS
- **Anthropic API Integration**: Go script for initial quote extraction using Claude API before custom model is ready
- **CI/CD Pipeline**: Automated testing and deployment via GitHub Actions
- **Enhanced Monitoring**: Custom Grafana dashboards for application metrics, alerting rules

---

## Security

Security was a priority throughout development:

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

**Problem**: Devices on the local network couldn't access the application via its public domain because traffic would route through the router and fail to return properly (hairpin NAT limitation).

**Solution**: Configured K3s CoreDNS with NodeHosts to resolve the domain directly to the Traefik ingress ClusterIP for in-cluster traffic. This bypasses the public IP entirely for requests originating within the cluster or LAN.

### Kubernetes Deployment Strategy

**Challenge**: Deploying a stateful application with external database dependencies on a minimal 2-node K3s cluster.

**Approach**:
- Used K3s `local-path` provisioner for persistent volumes (acceptable for homelab, node-local storage)
- PostgreSQL runs bare-metal on the worker node for performance; application connects via static LAN IP
- Traefik handles ingress with automatic TLS via cert-manager's HTTP-01 challenge
- Helm-managed monitoring stack (kube-prometheus-stack) with custom values for K3s compatibility
- Disabled unreachable scrape targets (kubeControllerManager, kubeScheduler, kubeEtcd, kubeProxy) since K3s bundles these into the main binary

### Session-Based Authentication

**Challenge**: Implementing secure, scalable authentication without JWT complexity.

**Solution**:
- PostgreSQL-backed session store via `gin-contrib/sessions`
- Sessions survive server restarts and scale across replicas
- Middleware chain: `AuthRequired()` → `QuoteOwnershipRequired(db)` for protected routes
- Auto-login after signup and password reset for seamless UX
- Username OR email login flexibility

### Kernel Module Compatibility (Arch Linux)

**Problem**: After an Arch Linux kernel update (6.17.8 → 6.18.7), kube-proxy failed with "Extension statistic revision 0 not supported, missing kernel module?" because the running kernel didn't match installed modules.

**Solution**: Identified the mismatch by comparing `uname -r` with installed kernel packages, then rebooted to load the correct modules. This is a common gotcha with rolling-release distributions in production.

---

## Monitoring Stack

Recently deployed kube-prometheus-stack for cluster observability:

- **Prometheus**: 15-day retention, 20Gi storage, scrapes all ServiceMonitors
- **Grafana**: Available at [grafana.zachmonroe.cloud](https://grafana.zachmonroe.cloud) with TLS
- **node-exporter**: DaemonSet collecting host metrics from both nodes
- **postgres-exporter**: Custom deployment scraping PostgreSQL metrics from the database server
- **Alertmanager**: Configured for future alerting rules

Dashboard configuration and custom alerting rules are planned for the coming weeks.

---

## AI-Assisted Development

This project was developed collaboratively with Claude Code (Anthropic's CLI tool), demonstrating effective human-AI pair programming patterns:

### Development Workflow

1. **Human-Written Foundation**: The core API structure, database schema, and frontend templates were written manually to establish architectural direction and coding style.

2. **Iterative Enhancement**: Claude Code was then used to extend functionality with specific, detailed prompts. For example:
   - "Add password reset functionality with email verification, 1-hour token expiration, and single-use tokens"
   - "Implement FLIP animations for card repositioning during hover expansion"

3. **Test-Driven Iteration**: Each feature was tested manually after implementation, with feedback provided for corrections. This caught edge cases and ensured the AI understood the codebase context.

4. **Security Coaching**: Explicit instructions were provided for handling sensitive data:
   - "Store credentials in environment variables, never commit to git"
   - "Add the secret.yaml to .gitignore before creating it"
   - "Use bcrypt with cost 12, not the default"

### Custom Deployment Agent

A custom `homelab-health-monitor` agent was configured to minimize friction during Kubernetes deployment:

- SSH access patterns pre-configured for the cluster nodes
- Sudo password handling automated (while keeping credentials out of prompts)
- Common kubectl commands wrapped with proper authentication
- Health checks for pods, nodes, certificates, and ingress resources

This reduced deployment time by eliminating repetitive context-setting and allowing focus on actual infrastructure decisions.

### Lessons Learned

- **Specific prompts yield better results**: "Add a logout button" produces inconsistent results; "Add a logout button to the header nav, POST to /auth/logout, clear the session, redirect to home" produces exactly what's needed.
- **Context is expensive**: Providing CLAUDE.md with project conventions prevented style drift and reduced corrections.
- **AI excels at boilerplate**: Kubernetes manifests, CRUD handlers, and repetitive patterns are ideal for AI generation.
- **Humans excel at architecture**: Database schema design, security decisions, and UX flow required human judgment.

---

## Local Development

```bash
# Clone the repository
git clone https://github.com/yourusername/zetl.git
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

Visit [zetl.zachmonroe.cloud](https://zetl.zachmonroe.cloud) to see the application in action:

1. Create an account (no email verification required for signup)
2. Add your favorite quotes with author, book, and tags
3. Click cards to flip and reveal details
4. Hover over cards to see the expansion animation
5. Use the tag filter to find specific quotes
6. Customize your profile and privacy settings

Feedback and suggestions are welcome!

---

## License

MIT License - see LICENSE file for details.

---

## Contact

Created by Zach Monroe

- GitHub: [github.com/zach-monroe](https://github.com/zach-monroe)
