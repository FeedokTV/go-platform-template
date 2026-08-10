[![CI](https://github.com/FeedokTV/go-platform-template/actions/workflows/ci.yml/badge.svg)](https://github.com/Feedok/go-platform-template/actions/workflows/ci.yml)

# Go Platform Template

When you starting your Go project as regular platform/application usually everybody do the same things from project to project. Me as well. That why I created this template firstly for myself and then for people who want to realise how platforms and (almost) production-grade application structure looks like. For me it was years of doing same thing from project to project and polishing it, now I want to share how I do it for me.

## Architecture

### Description
Hexagonal Architecture - A widely used approach for organising projects in programming. This project reveals this approach in lite version, so everybody can adapt it. 

Main advantage is - keep business logic from adapters, so we can easily **adapt** transport layers for main business logic. This approach helps us to make tests for components easily as well.

For many people this approach seems too "boring", but it give us guarantees and simple reintegration for various components (easily change DB/transport layer or etc).

Easiest example: The core (and it's models) defines a Repository port - interface that describing the persistence operations the domain needs (Create, Get, Update and etc.), while the service implements the specific business logic and then calls only that port, **never a driver**. Concrete adapters (PostgreSQL, Redis, etc.) implement the port: they translate domain models to storage format and driver errors into domain error kinds.Because the service depends on the interface alone, swapping the database means writing one new adapter package - core and port don't change.

### What this template deliberately is not

- No auth (orthogonal to the architecture)
- No gRPC (only one transport layer for template and for show the idea, second one as exercise for reader)
- No Kafka (no event volume for it)
- No DI framework (ADR-0004)
- No ORM (ADR-0002)
- No deploy target (template, where should it be deployed?)

### Parts

**Core (domain)** - concrete models (`Widget` in this project as example of model) + invariants
**Port** - interface the core demands (`widget.Repository`)
**Service** - business orchestration, consumes the port, knows **no** infrastructure
**Adapter** - implements the port (`postgresadapter`) or drives the core from outside (`httpserver`) 

How they acts between each other? 
- The domain service wants to persist data and calls a method on the port interface (WidgetRepository)
- The core knows nothing about databases; it depends solely on the port
- The adapter implements this port, converts domain models into the database format and executes the query
- When switching databases, only the adapter changes, while the core and the port remain the same

More in `ADR-0001`

More about hexagonal architecture: 
- https://en.wikipedia.org/wiki/Hexagonal_architecture_(software)
- https://docs.aws.amazon.com/prescriptive-guidance/latest/cloud-design-patterns/hexagonal-architecture.html

### Operational layer

- The application starts with validating config and attempt to ping the database. Application cannot run with unset required variables, but can run with incorrect DB credentials, but in unready state. `/healthz` endpoint shows that application started successfully, and `/readyz` that application ready to operate (database connection established), but if it's not the `/readyz` endpoint returns `"status": "unavailable"` (self-heal if connection established)
- On SIGTERM raised the application starting to close connections, close pools, finish the work gracefully.
- Down migrations exist as a dev convenience (`make migrate-down`), but the production problems roll forward with a new migration

#### Observability
- RED (Rate, Errors, Duration) metrics for measuring the performance of software services directly from the user's point of view, through `/metrics` endpoint with `observabilityMiddleware`. This middleware do logging as well. Each request has own `requestID` for correlation.
- Each endpoint covered in panic recovery middleware (`recoverMiddleware`)

#### Tests
- Integration tests is provided by `testcontainers` package and helps find the problems with database in current template. 
- Two types of tests: short (only unit-tests) and long (with integration tests)

#### Containerization
- Image is distroless to compact the size of it (23MB) 

#### CI
- Five CI jobs on every push: lint (golangci-lint v2), unit tests with -race, integration tests against real Postgres (testcontainers), vulnerability scan (govulncheck), and the Docker image build. Each maps to a Make target you can run locally first

## Quickstart

*Requires: Docker, Go 1.25+, make (Windows: WSL2 recommended)*

1. Prepare environment variables (contract of variables in `.env.example`)
`cp .env.example .env`

2. Run the docker compose for database (via Make for development environment)
`make up`

3. Run the migrations
`make migrate-up`

4. Run the application
`make run`

5. Curl it!

```
curl -X POST localhost:8080/v1/widgets -H "Content-Type: application/json" -d '{"name":"my widget","weight":5}'
# -> 201 {"id":1,"name":"my widget","weight":5,"created_at":...,"updated_at":...}

curl localhost:8080/readyz
# -> {"status":"ok"}
```

### List of endpoints

- `GET /healthz`- application state (alive or not)
- `GET /readyz` - application readiness for work (DB connection established)
- `GET /metrics` - RED metrics
- `POST /v1/widgets` - create widget
- `GET /v1/widgets/{id}` - fetch widget
- `PUT /v1/widgets/{id}` - update widget
- `DELETE /v1/widgets/{id}` - delete widget
- `GET /v1/widgets` - fetch all widgets

## Structure for project:
- `cmd/main.go` - main entrypoint
- `internal/core` - domain, models, business logic
- `internal/core/apperror` - shared errors for adapters (see in ADR-0005)
- `internal/platform` - everything that we need for platform like config, logger. Our instruments
- `internal/adapter` - adapters for ports like `httpserver` or `postgresadapter`. Concrete realisation for database or transportation
- `migrations` - PostgreSQL migrations
-  `deploy` - everything for deploy
- `.env.example` - contract for environment variables
- `.goreleaser.yaml` - release packaging: multi-platform archives, checksums, changelog. Builds live in make build and the Dockerfile

### Used packages:
Mostly I tried to use default packages (so everybody can easily adapt everything for their preferred libraries), its also a nice way to keep your application away from supply chain attacks

Packages:
- pgx/v5 for PostgreSQL driver
- testcontainers for integration tests

## How to adapt

For adapting this project, the developer should start with `core`: rename domain, rip out unnecessary pieces, and prepare the adapters

- Building of project goes through `make build`
- Updating the `depguard` rules in linter
- Extending the `.env.example`
- DTO's live in the adapter

## ADR's (Architectural Decision Records)

All decisions that made for the template and why. They all in `docs/adr`

- ADR-0001: Hexagonal over full Clean Architecture
- ADR-0002: pgx over GORM
- ADR-0003: slog over zap/zerolog
- ADR-0004: No DI framework
- ADR-0005: Error mapping at the boundary
- ADR-0006: stdlib net/http over chi/gin
- ADR-0007: Env-only config, stdlib over viper/koanf
- ADR-0008: Boot, shutdown, and dependency posture
- ADR-0009: Example-domain conventions (Widget)