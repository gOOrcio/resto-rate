# Resto Rate

Restaurant rating and review application built with SvelteKit frontend and Fastify API backend.

## Architecture

- **Frontend**: SvelteKit with TailwindCSS, Google OAuth authentication
- **Backend**: Fastify API with MessagePack, Google OAuth integration
- **Database**: PostgreSQL with shared schema
- **Communication**: MessagePack for efficient binary serialization
- **Authentication**: Google OAuth with secure session management
- **Monorepo**: Turborepo with bun package manager

## Quick Start

1. **Install dependencies**:

```bash
bun install
```

2. **Set up environment**:

```bash
cp env.template .env
# Edit .env with your database credentials and Google OAuth settings
```

3. **Set up Google OAuth** (required for authentication):

   a. Go to [Google Cloud Console](https://console.cloud.google.com/)
   b. Create a new project or select existing
   c. Enable Google+ API
   d. Create OAuth 2.0 credentials
   e. Add authorized redirect URIs:
      - Development: `http://localhost:3001/api/auth/google/callback`
      - Production: `https://yourdomain.com/api/auth/google/callback`
   f. Update your `.env` file with the credentials

4. **Start database** (Docker recommended):

```bash
cd apps/web && bun run db:start
```

5. **Run migrations**:

```bash
# Run the Google OAuth migration
cd apps/api && bun run db:migrate-google-oauth

# Push schema changes
cd apps/web && bun run db:push
```

6. **Start development**:

```bash
bun run dev
```

This starts both the web app (http://localhost:5173) and API (http://localhost:3001).

## Authentication

The app uses Google OAuth for authentication:

- **No manual user registration** - Users sign in with Google
- **Secure sessions** - 30-day session duration with automatic expiry
- **Protected routes** - Frontend guards for authenticated pages
- **User data** - Only stores email and name from Google

## Project Structure

```
resto-rate/
├── apps/
│   ├── web/          # SvelteKit frontend
│   └── api/          # Fastify backend
├── packages/
│   ├── config/       # Shared environment config
│   ├── database/     # Database schema and types
│   ├── constants/    # Shared constants and types
│   └── logger/       # Logging utilities
├── .env              # Environment variables
└── env.template      # Environment template
```

## Development

- `bun run dev` - Start both apps
- `bun run build` - Build all apps
- `bun run lint` - Lint all packages
- `cd apps/api && bun run db:migrate-google-oauth` - Run Google OAuth migration

## Documentation

See `SETUP.md` for detailed setup instructions and architecture overview.
