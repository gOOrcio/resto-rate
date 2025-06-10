# Resto Rate

Restaurant rating and review application built with SvelteKit frontend and Fastify API backend.

## Architecture

- **Frontend**: SvelteKit with TailwindCSS, Lucia auth, Drizzle ORM
- **Backend**: Fastify API with MessagePack, authentication middleware
- **Database**: PostgreSQL with shared schema
- **Communication**: MessagePack for efficient binary serialization
- **Monorepo**: Turborepo with bun package manager

## Quick Start

1. **Install dependencies**:

```bash
bun install
```

2. **Set up environment**:

```bash
cp env.template .env
# Edit .env with your database credentials
```

3. **Start database** (Docker recommended):

```bash
cd apps/web && bun run db:start
```

4. **Run migrations**:

```bash
cd apps/web && bun run db:push
```

5. **Start development**:

```bash
bun run dev
```

This starts both the web app (http://localhost:5173) and API (http://localhost:3001).

## Project Structure

```
resto-rate/
├── apps/
│   ├── web/          # SvelteKit frontend
│   └── api/          # Fastify backend
├── packages/
│   └── config/       # Shared environment config
├── .env              # Environment variables
└── env.template      # Environment template
```

## Development

- `bun run dev` - Start both apps
- `bun run build` - Build all apps
- `bun run lint` - Lint all packages

## Documentation

See `SETUP.md` for detailed setup instructions and architecture overview.
