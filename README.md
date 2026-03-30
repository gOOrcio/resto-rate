# Resto Rate

Social restaurant rating app — find restaurants via Google Places, rate them, build a wishlist, and share with friends.

## Architecture

- **Frontend**: SvelteKit 5 (Svelte runes) + TailwindCSS + ShadCN Svelte + Paraglide i18n
- **Backend**: Go + Connect-RPC (gRPC-compatible) + PostgreSQL + Valkey cache
- **API contract**: Protocol Buffers (source of truth in `packages/protos/`)
- **Monorepo**: Nx workspace managed with bun
- **Auth**: Session-based (cookie + Valkey); Google OAuth planned

## Current Features

### Implemented
- **All 5 Connect-RPC services**: `auth`, `users`, `restaurants`, `reviews`, `google_maps`
- **Auth**: Session login (username), logout, get current user — sessions stored in Valkey (24h TTL, HttpOnly cookie)
- **Restaurants**: Full CRUD + paginated list; created from Google Places data (GoogleID + address as unique identifiers)
- **Reviews**: Create (with find-or-create restaurant), get, update, delete, list — one review per user per restaurant enforced at DB level; supports 1–5 star rating, comment, and free-form tags (stored as JSON array)
- **Google Places**: Text search, autocomplete (session-token batching), get place details, search restaurants, get restaurant details — comprehensive `Place` proto with 100+ fields
- **UI**: Restaurant search (autocomplete), restaurant card (inline edit + Google details panel), rating form (stars + comment + tags), review summary, login modal

### In Progress / Planned (MVP)
- Google OAuth login (replace username-only flow)
- My reviews page (list + filter + sort)
- Wishlist (save restaurants without rating)
- Friends (friend/unfriend, browse friend reviews + wishlist)

## Quick Start

1. **Install dependencies**:
```bash
bun install
```

2. **Start infrastructure** (PostgreSQL + Valkey):
```bash
docker compose up -d
```

3. **Set up environment**:
```bash
cp env.template .env
# Edit .env — set POSTGRES_* and GOOGLE_PLACES_API_KEY at minimum
```

4. **Start development**:
```bash
bun run dev
```

Web: http://localhost:5173 · API: http://localhost:3001

## Proto → Code Generation

All API contracts live in `packages/protos/`. To regenerate after editing `.proto` files:

```bash
nx run protos:generate       # both Go and TypeScript
nx run protos:generate:api   # Go only
nx run protos:generate:web   # TypeScript only
```

## Database Schema

### Users
| Column | Type | Notes |
|--------|------|-------|
| id | UUIDv7 | PK |
| google_id | string | unique, nullable |
| email | string | unique, nullable |
| username | string | unique |
| name | string | |
| is_admin | bool | |

### Restaurants
| Column | Type | Notes |
|--------|------|-------|
| id | UUIDv7 | PK |
| google_id | string | unique — full Places resource name (`places/ChIJ...`) |
| name | string | |
| address | string | unique |

### Reviews
| Column | Type | Notes |
|--------|------|-------|
| id | UUIDv7 | PK |
| user_id | string | FK → users |
| restaurant_id | string | FK → restaurants |
| google_places_id | string | indexed |
| rating | float64 | 1–5 |
| comment | string | |
| tags | JSON array | free-form strings |
| (restaurant_id, user_id) | composite unique | one review per user per restaurant |

## Development Commands

```bash
bun run dev          # Start web + API concurrently
bun run build        # Build all
bun run lint         # Lint all
bun run test         # Run all tests
bun run format       # Prettier

cd apps/api && go build ./...       # Verify Go compiles
cd apps/web && bun run check        # svelte-check
```

## Tech Stack

| Layer | Technology |
|-------|-----------|
| Frontend | SvelteKit 5, Svelte 5 runes, TailwindCSS v4, ShadCN Svelte, Lucide |
| API client | Connect-RPC (TypeScript), generated from protos |
| Backend | Go, Connect-RPC, GORM |
| Database | PostgreSQL (UUIDv7 PKs, GORM auto-migrate) |
| Cache | Valkey (Redis-compatible) — sessions + proto caching |
| API contract | Protocol Buffers + Buf CLI |
| Monorepo | Nx + bun |
| Observability | Prometheus metrics, structured slog logging |
| i18n | Paraglide |

See `ROADMAP.md` for the MVP implementation plan.
