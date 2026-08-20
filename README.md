# autobee-server

> Production-ready Go + MongoDB backend boilerplate for a multi-tenant vehicle service management platform.

---

## Table of Contents

- [Project Overview](#project-overview)
- [Architecture](#architecture)
- [Tech Stack](#tech-stack)
- [Requirements](#requirements)
- [Installation](#installation)
- [Environment Variables](#environment-variables)
- [Running Locally](#running-locally)
- [Docker Setup](#docker-setup)
- [Running Tests](#running-tests)
- [Code Coverage](#code-coverage)
- [Linting](#linting)
- [Formatting](#formatting)
- [Swagger](#swagger)
- [Seed Data](#seed-data)
- [API Versioning](#api-versioning)
- [Authentication](#authentication)
- [RBAC](#rbac)
- [Multi-Tenancy](#multi-tenancy)
- [CORS](#cors)
- [Project Structure](#project-structure)
- [Git Conventions](#git-conventions)

---

## Project Overview

`autobee-server` is a clean, extensible backend boilerplate for a multi-tenant vehicle service management platform. It provides:

- JWT-based authentication (access + refresh tokens)
- Role-based access control (super_admin, tenant_admin, user)
- Server-side tenant isolation
- Reference vehicle module demonstrating the full request lifecycle
- Stub modules for bookings, repairs, service centers, and services
- Production-grade logging, error handling, and graceful shutdown
- Full Docker and CI setup out of the box

---

## Architecture

```
HTTP Request
    ↓
Middleware (CORS → RequestID → Logger → Auth → TenantContext)
    ↓
Handler     (parse, validate, call service, format response)
    ↓
Service     (business logic, authorization, tenant isolation)
    ↓
Repository  (MongoDB queries only — no business logic)
    ↓
MongoDB
```

---

## Tech Stack

| Layer | Technology |
|---|---|
| Language | Go 1.23+ |
| Router | Chi v5 |
| Database | MongoDB (official driver v2) |
| Auth | JWT (`golang-jwt/jwt/v5`) + bcrypt |
| Validation | go-playground/validator v10 |
| Logging | Zap |
| Testing | Testify + Testcontainers |
| API Docs | Swaggo |
| Linting | golangci-lint |
| Git Hooks | Lefthook |
| Containerization | Docker + Docker Compose |
| CI | GitHub Actions |

---

## Requirements

- Go 1.23+
- Docker (for Testcontainers integration tests and docker-compose)
- `golangci-lint` (install via `make tools`)
- `swag` (install via `make tools`)
- `goimports` (install via `make tools`)

---

## Installation

```bash
git clone https://github.com/autobee/server.git
cd server

# Install dev tools
make tools

# Copy environment configuration
cp .env.example .env
# Edit .env and fill in your secrets

# Download Go dependencies
go mod download
```

---

## Environment Variables

| Variable | Required | Description |
|---|---|---|
| `APP_ENV` | No | `development` \| `production` (default: `development`) |
| `APP_NAME` | No | Application name used in logs |
| `PORT` | No | HTTP port (default: `8080`) |
| `MONGO_URI` | **Yes** | MongoDB connection string |
| `MONGO_DATABASE` | **Yes** | MongoDB database name |
| `JWT_ACCESS_SECRET` | **Yes** | Access token signing secret (≥32 chars) |
| `JWT_REFRESH_SECRET` | **Yes** | Refresh token signing secret (≥32 chars) |
| `JWT_ACCESS_EXPIRATION` | No | Access token TTL (default: `15m`) |
| `JWT_REFRESH_EXPIRATION` | No | Refresh token TTL (default: `168h`) |
| `CORS_ALLOWED_ORIGINS` | No | Comma-separated allowed origins |
| `CORS_ALLOWED_METHODS` | No | Comma-separated allowed methods |
| `CORS_ALLOWED_HEADERS` | No | Comma-separated allowed headers |
| `CORS_ALLOW_CREDENTIALS` | No | `true` \| `false` (default: `true`) |
| `LOG_LEVEL` | No | `debug` \| `info` \| `warn` \| `error` |

Generate strong JWT secrets:

```bash
openssl rand -hex 32
```

---

## Running Locally

### Option 1 — Local MongoDB

```bash
# Start MongoDB (or use docker-compose)
make docker-up

# Start the API server
make dev
```

### Option 2 — Full Docker

```bash
# Build and start all services
make docker-up

# API is available at http://localhost:8080
```

### Verify it works

```bash
curl http://localhost:8080/health
# {"data":{"status":"ok"}}

curl http://localhost:8080/ready
# {"data":{"status":"ready"}}
```

---

## Docker Setup

```bash
# Start MongoDB + API containers
make docker-up

# Stop containers
make docker-down

# Stop and remove volumes (⚠️ deletes data)
make docker-clean
```

MongoDB data is persisted in the `autobee-mongo-data` named volume.

---

## Running Tests

```bash
# Unit tests only (fast, no Docker needed)
make test

# Integration tests (requires Docker for Testcontainers)
make test-integration

# All tests
make test-coverage
```

Integration tests start a real MongoDB container automatically via Testcontainers. No manual MongoDB setup required.

---

## Code Coverage

```bash
make test-coverage
# Generates: coverage.out + coverage.html
# Opens: coverage.html in browser

open coverage.html
```

---

## Linting

```bash
make lint
```

Linters enabled: `errcheck`, `govet`, `staticcheck`, `ineffassign`, `unused`, `gosimple`, `gosec`, `revive`, `bodyclose`, `contextcheck`, `gofmt`, `goimports`, `misspell`.

---

## Formatting

```bash
# Auto-format all files
make format

# Check only (used in CI)
make format-check
```

---

## Swagger

```bash
# Generate docs (must be done before first run)
make swagger

# Start the server then visit:
open http://localhost:8080/api/docs/
```

Swagger supports JWT Bearer authentication. Click **Authorize**, paste your access token, and test protected endpoints directly in the UI.

---

## Seed Data

```bash
# Requires .env with valid MONGO_URI and MONGO_DATABASE
make seed
```

Creates the following development records (idempotent):

| Role | Phone | Password |
|---|---|---|
| Super Admin | `+910000000001` | `SuperAdmin@123` |
| Tenant Admin | `+910000000002` | `TenantAdmin@123` |
| User | `+910000000003` | `User@1234` |

The script also prints the **Tenant ID** — copy it for use in signup/signin requests.

> ⚠️ **Never use seed credentials in production.**

---

## API Versioning

All endpoints are prefixed with `/api/v1`:

```
POST /api/v1/auth/signup
POST /api/v1/auth/signin
POST /api/v1/auth/refresh

GET  /api/v1/users/me

GET  /api/v1/tenants
POST /api/v1/tenants
GET  /api/v1/tenants/{id}

POST /api/v1/vehicles
GET  /api/v1/vehicles
GET  /api/v1/vehicles/{id}
```

---

## Authentication

### Signup

```bash
POST /api/v1/auth/signup
Content-Type: application/json

{
  "name": "John Doe",
  "phone": "+911234567890",
  "password": "securepassword123",
  "tenantId": "<tenant-object-id>"
}
```

### Signin

```bash
POST /api/v1/auth/signin
Content-Type: application/json

{
  "phone": "+911234567890",
  "password": "securepassword123",
  "tenantId": "<tenant-object-id>"
}
```

### Using the Access Token

```bash
GET /api/v1/users/me
Authorization: Bearer <access-token>
```

### Refresh Token

```bash
POST /api/v1/auth/refresh
Content-Type: application/json

{
  "refreshToken": "<refresh-token>"
}
```

Token lifetimes: access = 15 minutes, refresh = 7 days (configurable via env).

---

## RBAC

Three roles are defined:

| Role | Access |
|---|---|
| `super_admin` | All tenants, all resources |
| `tenant_admin` | All resources within their tenant |
| `user` | Only their own resources |

RBAC is enforced via Chi middleware:

```go
r.Use(middleware.RequireRole(auth.RoleSuperAdmin))
r.Use(middleware.RequireRole(auth.RoleSuperAdmin, auth.RoleTenantAdmin))
```

---

## Multi-Tenancy

Tenant context is derived **exclusively from the authenticated user's JWT**. Clients cannot inject a `tenantId` through request parameters.

Tenant isolation is enforced in the service layer:

- `RoleUser` → sees only their own resources within their tenant
- `RoleTenantAdmin` → sees all resources within their tenant
- `RoleSuperAdmin` → sees resources across all tenants

Cross-tenant access returns `403 Forbidden`.

---

## CORS

CORS is configured from environment variables and applied globally before all routes:

```env
CORS_ALLOWED_ORIGINS=http://localhost:3000,http://localhost:5173
CORS_ALLOWED_METHODS=GET,POST,PUT,PATCH,DELETE,OPTIONS
CORS_ALLOWED_HEADERS=Accept,Authorization,Content-Type,X-Request-ID
CORS_ALLOW_CREDENTIALS=true
```

> **Important:** When `CORS_ALLOW_CREDENTIALS=true`, wildcard (`*`) origins are not permitted per the CORS specification.

---

## Project Structure

```
autobee-server/
├── cmd/api/main.go          # Entry point, dependency wiring
├── internal/
│   ├── auth/                # Authentication (JWT, signup, signin, refresh)
│   ├── users/               # User profile (/users/me)
│   ├── tenants/             # Tenant management (super_admin only)
│   ├── vehicles/            # Reference module with full tenant isolation
│   ├── bookings/            # Stub — implement following vehicles pattern
│   ├── repairs/             # Stub
│   ├── service_centers/     # Stub
│   ├── services/            # Stub
│   ├── config/              # Centralized env configuration
│   ├── database/            # MongoDB client + index management
│   ├── middleware/          # Auth, CORS, Logger, RequestID, Tenant
│   └── server/              # HTTP server + route registration
├── pkg/
│   ├── response/            # Standardized JSON response helpers
│   ├── logger/              # Zap logger factory
│   ├── validator/           # Singleton validator with formatted errors
│   └── password/            # bcrypt Hash + Compare
├── tests/integration/       # Testcontainers E2E tests
├── scripts/seed.go          # Development seed data
├── docs/                    # Swagger generated output
├── .github/workflows/ci.yml # GitHub Actions CI pipeline
├── Dockerfile               # Multi-stage Docker build
├── docker-compose.yml       # Local dev infrastructure
├── Makefile                 # Developer commands
├── .golangci.yml            # Linter configuration
├── .lefthook.yml            # Git hooks
└── .env.example             # Environment variable template
```

---

## Git Conventions

Follow [Conventional Commits](https://www.conventionalcommits.org/):

```
feat: add vehicle module
fix: prevent cross-tenant access
test: add auth service integration tests
refactor: simplify JWT middleware
docs: update local setup instructions
chore: update dependencies
ci: add coverage upload step
```

Enforced by Lefthook commit-msg hook.

Install Lefthook:

```bash
brew install lefthook  # macOS
lefthook install       # installs git hooks
```

---

## Security Notes

- Passwords are bcrypt hashed (cost 12)
- JWT secrets must be ≥32 chars and stored in `.env` (never committed)
- Authorization headers, passwords, OTPs, and tokens are never logged
- MongoDB errors are never exposed to API clients
- All input is validated at the HTTP boundary
- Request body size is limited to 1 MB
- All DB operations use context-aware calls with timeouts
- Non-root user in Docker runtime image
