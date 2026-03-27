# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Commands

### Development

```bash
bun run dev          # Start both web (port 5173) and API (port 3001) concurrently
bun run dev:web      # Start only the SvelteKit frontend
bun run dev:api      # Start only the Go API (uses `air` for hot reload)
```

### Build, Lint, Test

Always use Nx to build, test, and lint — never invoke `go build` or `bun run check` directly.

There is no global `nx`. Use `bunx nx` or `./node_modules/.bin/nx` — both run the workspace-local Nx from `node_modules`.

```bash
# Per-app Nx targets (preferred for targeted checks)
bunx nx run api:build     # Build Go API
bunx nx run api:test      # Run Go tests
bunx nx run api:lint      # Lint Go code
bunx nx run web:build     # Build SvelteKit frontend
bunx nx run web:test      # Run frontend tests
bunx nx run web:lint      # Lint frontend
bunx nx run web:check     # svelte-check TypeScript + Svelte types

# Convenience wrappers (bun run X calls nx under the hood)
bun run build        # Build all apps
bun run lint         # Lint all packages
bun run check-types  # TypeScript type checking
bun run test         # Run all tests
bun run format       # Prettier format all files

# Nx affected (only run on changed packages)
bun run test:affected
bun run lint:affected
bun run build:affected
```

### Running a single Go test

```bash
cd apps/api && go test ./src/test/... -run TestName -v
```

### Quick verification (without running tests)

```bash
# Use Nx targets, not raw go/bun commands:
bunx nx run api:build     # Verify Go compiles
bunx nx run web:check     # svelte-check TypeScript + Svelte types
```

### Protobuf code generation

```bash
bunx nx run protos:generate       # Generate both Go and TypeScript from .proto files
bunx nx run protos:generate:api   # Go only
bunx nx run protos:generate:web   # TypeScript only
bunx nx run protos:clean          # Remove all generated files
```

### Infrastructure

```bash
docker-compose up -d         # Start PostgreSQL + Valkey (Redis-compatible cache)
docker-compose down          # Stop services
```

## Architecture

This is an **Nx monorepo** managed with **bun**, containing:

```
apps/
  api/      # Go backend
  web/      # SvelteKit frontend
packages/
  protos/   # Protobuf definitions (source of truth for API contracts)
```

### API Contract: Protocol Buffers + Connect-RPC

All API contracts are defined in `packages/protos/` (organized by service: `restaurants/`, `users/`, `google_maps/`, `reviews/`). Code generation outputs to:
- Go: `apps/api/src/generated/`
- TypeScript: `apps/web/src/lib/client/generated/`

The frontend calls the API using the generated Connect-RPC clients (`apps/web/src/lib/client/client.ts`), which expose typed `restaurants`, `users`, `googleMaps`, and `reviews` clients.

**When adding/changing API endpoints:** edit `.proto` files in `packages/protos/`, run `nx run protos:generate`, then implement the handler in `apps/api/src/services/`.

**Current services:** `restaurants`, `users`, `google_maps`, `auth`, and `reviews` are all fully implemented. All 5 services are registered in `main.go` with Prometheus metrics interceptors.

**Go service implementation pattern:** Each service embeds the generated `v1connect.UnimplementedXxxServiceHandler`, holds a `*gorm.DB` (or other client), and is registered in `src/main.go` via `initializeServiceHandlers()` with a Prometheus metrics interceptor.

### Go API (`apps/api`)

- Entry point: `src/main.go` — wires DB, cache, services, HTTP mux, CORS middleware, Prometheus metrics, and gRPC reflection (dev only)
- `src/services/` — Connect-RPC service implementations (`RestaurantsService`, `UsersService`, `GooglePlacesAPIService`, `AuthService`, `ReviewsService`)
- `src/internal/models/` — GORM models with UUIDv7 primary keys (`uuid7_model.go` is the base model)
- `src/internal/mappers/` — conversion between GORM models and protobuf types
- `src/internal/cache/` — Valkey client and proto caching utilities
- `src/internal/utils/` — DB schema auto-migration, seeding, logging, pagination, Prometheus metrics
- Hot reload in dev via `air` (config: `apps/api/.air.toml`)
- In `dev` mode (`ENV=dev`), runs HTTP (h2c); in production runs HTTPS with cert.pem/key.pem

### SvelteKit Frontend (`apps/web`)

- Uses **Svelte 5 runes**: `$state()`, `$props<Type>()`, `$derived()` — not legacy stores or `export let`
- UI components: **ShadCN Svelte** (headless Tailwind components in `src/lib/components/ui/`); icons from `lucide-svelte`
- i18n: **Paraglide** (`src/lib/paraglide/`) with server middleware in `src/hooks.server.ts` and URL rewriting in `src/hooks.ts`
- API URL configured via `VITE_API_URL` env var (defaults to `http://localhost:3001`)
- Builds output to `dist/apps/web`

### Environment Variables

The root `.env` file is used by the API. Key variables:
- `ENV` — `dev` or production (controls HTTP vs HTTPS, CORS, gRPC reflection)
- `POSTGRES_*` — database connection (host, user, password, db, port)
- `VALKEY_URI`, `VALKEY_PASSWORD` — cache connection
- `API_PORT`, `API_HOST`, `API_PROTOCOL`, `WEB_UI_PORT` — server config
- `SEED=true` — enables database seeding on startup
- `LOG_LEVEL` — slog level (DEBUG/INFO/WARN/ERROR or numeric)
- `GOOGLE_PLACES_API_KEY` — required for Google Maps service
- `VITE_API_URL`, `VITE_PORT` — frontend-specific (in `apps/web/.env`)

### Database

- PostgreSQL with **GORM** (schema auto-migrated on startup via `utils.CreateSchema`)
- All models use UUIDv7 primary keys (via `uuid7_model.go` base model)
- Valkey (Redis-compatible) used for caching proto responses
