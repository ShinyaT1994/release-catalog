# Release Catalog

A Dependency-Track–integrated release catalog service. It manages a product's
Main/Release branch lifecycle and resolves SBOM composition by recursively
following CycloneDX **BOM-Link** references into a release dependency graph.

The system pairs a Go (Echo) backend with a React (Vite) frontend, and talks to
a [Dependency-Track](https://dependencytrack.org/) v5 instance (or a built-in
stub) for project and BOM data.

## Features

- **Product management** — CRUD for products.
- **Branch lines** — every product has a `MAIN` branch line and any number of
  `RELEASE` branch lines forked from a snapshot. Branches carry a lifecycle
  status (`active`, `maintenance`, `security_only`, `end_of_support`, `closed`).
- **Current state & snapshots** — each branch tracks a mutable *current state*
  (root DT project UUID, root BOM serial/version/SHA-256, source revision) and
  can produce immutable snapshots: `MAIN_SNAPSHOT` and `RELEASE`.
- **BOM-Link graph resolution** — starting from a root DT project, the resolver
  parses `urn:cdx:` BOM-Link external references, resolves child projects
  recursively, detects cycles, and reports unresolved links. Traversal is
  bounded by `maxDepth` and `maxNodes`.
- **Dependency-Track adapter** — a pluggable client with an HTTP implementation
  and an in-memory stub for local development and tests.

## Architecture

The backend follows a package-by-feature, ports-and-adapters layout. Each
feature package (`product`, `branch`, `release`, `graph`) contains its own
domain, usecase, repository, handler, and ports. Cross-feature dependencies are
wired through adapters in `cmd/server/main.go`.

```
release-catalog/
├── cmd/server/            # Entry point + cross-feature adapter wiring
├── internal/
│   ├── product/           # Products
│   ├── branch/            # Branch lines + current state
│   ├── release/           # Snapshots (main snapshot / release)
│   ├── graph/             # BOM-Link resolver + release graph
│   ├── dtclient/          # Dependency-Track client (HTTP + stub)
│   ├── dtproxy/           # DT passthrough handler
│   └── shared/            # config, database (SQLite migrations),
│                          #   middleware, apperror
├── api/openapi.yaml       # OpenAPI 3.1 spec (source of truth)
├── frontend/              # React 19 + Vite + TanStack Query
├── testdata/boms/         # Sample CycloneDX BOMs
└── scripts/               # Seed scripts for test data
```

### Tech stack

| Layer            | Choice                                         |
|------------------|------------------------------------------------|
| Language         | Go 1.25                                        |
| Web framework    | Echo v4                                         |
| Database         | SQLite (`mattn/go-sqlite3`, CGo)               |
| Logging          | `slog` (JSON)                                   |
| API contract     | OpenAPI 3.1 (`api/openapi.yaml`)               |
| Frontend         | React 19, Vite, TanStack Query, React Router    |
| Container        | Docker / Docker Compose                         |

## Prerequisites

- Go 1.25+ with CGo enabled (a C toolchain is required for `go-sqlite3`).
- Node.js 20+ (for the frontend).
- Docker + Docker Compose (optional, for containerized run).

## Getting started

### 1. Configure

Copy the example environment file and adjust as needed:

```bash
cp .env.example .env
```

| Variable                 | Default                     | Description                                       |
|--------------------------|-----------------------------|---------------------------------------------------|
| `RC_DATABASE_DRIVER`     | `sqlite`                    | Database driver.                                  |
| `RC_DATABASE_DSN`        | `./release-catalog.db`      | SQLite file path.                                 |
| `RC_SERVER_PORT`         | `8080`                      | HTTP listen port.                                 |
| `RC_DT_BASE_URL`         | `http://localhost:8081`     | Dependency-Track API base URL.                    |
| `RC_DT_API_KEY`          | *(empty)*                   | Dependency-Track API key.                         |
| `RC_DT_STUB_MODE`        | `true`                      | Use the in-memory DT stub instead of a real DT.   |
| `RC_DT_TIMEOUT_SECONDS`  | `30`                        | DT HTTP client timeout.                           |
| `RC_LOG_LEVEL`           | `info`                      | `debug` / `info` / `warn` / `error`.              |
| `RC_CORS_ALLOW_ORIGINS`  | `*`                         | Comma-separated allowed CORS origins.             |

With `RC_DT_STUB_MODE=true` (the default) no real Dependency-Track instance is
required.

### 2. Run the backend

```bash
make run          # build + run
# or
make build        # produces bin/release-catalog
make test         # run all tests
make test-coverage
make lint         # golangci-lint
```

The server runs on `http://localhost:8080`. Migrations run automatically on
startup. Health check:

```bash
curl http://localhost:8080/health
```

### 3. Run the frontend

```bash
cd frontend
npm install
npm run dev       # Vite dev server
```

## API

All endpoints are served under `/api/v1`. The full contract is in
[`api/openapi.yaml`](api/openapi.yaml).

| Method & Path                                   | Description                          |
|-------------------------------------------------|--------------------------------------|
| `GET /health`                                   | Health check                         |
| `POST /api/v1/products`                         | Create a product                     |
| `GET /api/v1/products`                          | List products                        |
| `GET /api/v1/products/{productId}`              | Get a product                        |
| `PATCH /api/v1/products/{productId}`            | Update a product                     |
| `DELETE /api/v1/products/{productId}`           | Delete a product                     |
| `GET /api/v1/products/{productId}/branches`     | List branches for a product          |
| `POST /api/v1/products/{productId}/release-lines` | Create a release line              |
| `GET /api/v1/branches/{branchId}`               | Get a branch                         |
| `PATCH /api/v1/branches/{branchId}`             | Update a branch                      |
| `GET /api/v1/branches/{branchId}/current`       | Get branch current state             |
| `PUT /api/v1/branches/{branchId}/current`       | Update branch current state          |
| `POST /api/v1/branches/{branchId}/snapshots`    | Create a main snapshot               |
| `POST /api/v1/branches/{branchId}/releases`     | Create a release                     |
| `GET /api/v1/branches/{branchId}/releases`      | List releases for a branch           |
| `GET /api/v1/releases/{releaseId}`              | Get a release                        |
| `GET /api/v1/branches/{branchId}/current/graph` | Release graph for branch current state |
| `GET /api/v1/releases/{releaseId}/graph`        | Release graph for a release          |

Graph endpoints accept `maxDepth` (default 10) and `maxNodes` (default 1000)
query parameters. When Dependency-Track is unreachable, graph endpoints return
`502`.

## Data model

- **Product** — top-level entity.
- **BranchLine** — `MAIN` or `RELEASE`; a release line records
  `sourceBranchLineId` and `forkedFromSnapshotId`.
- **BranchCurrentState** — mutable per-branch pointer to the current root DT
  project / BOM.
- **Snapshot** — immutable `MAIN_SNAPSHOT` or `RELEASE` capture with lifecycle
  status.
- **BOM-Link index** — resolved edges between DT projects derived from
  CycloneDX `urn:cdx:` references.

## Docker

Run the backend + frontend together:

```bash
make docker-up      # docker compose up -d
make docker-down
```

The frontend is served on port `3000`, the backend on `8080`.

### Dependency-Track reference stack

A standalone Dependency-Track v5 stack is provided for integration testing (it
is **not** started with the app):

```bash
make dt-up          # docker compose -f docker-compose.dt.yml up -d
make dt-down
```

This starts the DT API (`8081`), DT frontend (`8082`), and PostgreSQL.

## Test data

Sample CycloneDX BOMs live in [`testdata/boms/`](testdata/boms/). Seed scripts
push a nested TopBOM → SubBOM structure into a running DT + Release Catalog:

```bash
make seed-top-sub   # requires DT_API_KEY and a running DT + RC
```

## License

MIT (per `api/openapi.yaml`).
