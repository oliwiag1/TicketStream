# TicketStream

TicketStream is a ticket reservation platform focused on anti-oversell flow under concurrent traffic.

## Core Flow

1. Login via Keycloak (OIDC Authorization Code + PKCE).
2. Fetch event list and seats.
3. Reserve seat with Redis lock + SQL state update.
4. Pay with idempotency key.
5. Publish outbox event to RabbitMQ.
6. Worker processes ticket job with retry and DLQ.
7. Frontend receives realtime seat status updates over WebSocket.

## Architecture

- Backend: Go + Echo
- Frontend: React + Vite
- Database: PostgreSQL
- Cache/locks: Redis
- Messaging: RabbitMQ
- Auth: Keycloak

See [docs/adr/0001-system-architecture.md](docs/adr/0001-system-architecture.md) for design decisions.
Full ADR index: [docs/adr/README.md](docs/adr/README.md).

## Quick Start (Docker Compose)

Prerequisites:

- Docker + Docker Compose

Run:

```bash
docker compose up --build
```

Endpoints:

- Frontend: http://localhost:3000
- Backend API: http://localhost:8080
- Keycloak: http://localhost:18081
- RabbitMQ UI: http://localhost:15672

## Local Development

Prerequisites:

- Go 1.22+
- Node.js 20+
- npm
- k6 (for load tests)

Backend:

```bash
cd backend
make tidy
make run-api
```

Worker:

```bash
cd backend
make run-worker
```

Frontend:

```bash
cd frontend
npm install
npm run dev
```

## Testing

Backend unit tests:

```bash
cd backend
make test-unit
```

Integration tests:

```bash
cd backend
make test-integration
```

E2E tests:

```bash
cd backend
make test-e2e
```

Flash sale load test (k6):

```bash
cd backend
LOAD_BASE_URL=http://localhost:8080 \
LOAD_BEARER_TOKEN=<TOKEN> \
LOAD_EVENT_ID=10000000-0000-0000-0000-000000000001 \
LOAD_SEAT_IDS=20000000-0000-0000-0000-000000000001,20000000-0000-0000-0000-000000000002 \
make test-load
```

## API Documentation

- OpenAPI spec: [docs/api/openapi.yaml](docs/api/openapi.yaml)
- Quick preview: open `docs/api/openapi.yaml` in Swagger Editor (https://editor.swagger.io)

## CI/CD

- GitHub Actions workflow: [.github/workflows/ci-cd.yml](.github/workflows/ci-cd.yml)
- On every push and pull request:
	- backend lint (`gofmt` + `go vet`)
	- backend tests (`go test ./...`)
	- frontend build verification (`npm run build`)
- On push to `main`:
	- build candidate Docker images (`docker compose build api frontend`)

## Requirements Coverage

- Professor requirements status: [docs/runbooks/requirements_coverage.md](docs/runbooks/requirements_coverage.md)

## Operational Endpoints

- Health: `GET /health`
- Liveness: `GET /health/live`
- Readiness: `GET /health/ready`
- Metrics (JSON): `GET /metrics`
- Dashboard (HTML): `GET /ops/dashboard`

## Security and Reliability

- Redis-based rate limiting for auth and reservation endpoints.
- CSRF origin/referer checks for mutating endpoints.
- Security headers middleware and frontend nginx headers.
- Structured request logs with request_id and trace_id.
- Background sweeper that releases expired reservations and broadcasts seat availability.

## Deployment Profiles

- Kubernetes profile: [deploy/k8s/profile.md](deploy/k8s/profile.md)
- ECS profile: [deploy/ecs/profile.md](deploy/ecs/profile.md)

## Infra and Runbooks

- PostgreSQL replica and failover plan: [docs/infra/postgres-replica-failover-plan.md](docs/infra/postgres-replica-failover-plan.md)
- Redis Sentinel/Cluster plan: [docs/infra/redis-sentinel-cluster-plan.md](docs/infra/redis-sentinel-cluster-plan.md)
- Backup/restore runbook: [docs/runbooks/backup_restore.md](docs/runbooks/backup_restore.md)
