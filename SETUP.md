# Resto Rate - Setup Guide

Complete setup guide for the Turbo monorepo with SvelteKit frontend and Fastify API backend.

## Architecture Overview

- **Frontend** (`apps/web`): SvelteKit with TailwindCSS, Lucia auth, Drizzle ORM
- **Backend** (`apps/api`): Fastify with MessagePack, Drizzle ORM, shared authentication
- **Database**: PostgreSQL with shared schema between frontend and backend
- **Communication**: MessagePack for efficient binary serialization

## Prerequisites

- Node.js 18+
- Bun (package manager)
- PostgreSQL
- Docker (optional, for database)

## Quick Setup

1. **Install dependencies**:

```bash
bun install
```

2. **Database setup**:

```bash
# Option 1: Using Docker (recommended)
cd apps/web && bun run db:start

# Option 2: Local PostgreSQL
# Create database manually and configure connection string
```

3. **Environment configuration**:

**For Web App** (`apps/web/.env`):

```env
DATABASE_URL="postgresql://username:password@localhost:5432/resto_rate"
```

**For API** (`apps/api/.env`):

```env
DATABASE_URL="postgresql://username:password@localhost:5432/resto_rate"
PORT=3001
NODE_ENV=development
```

4. **Run database migrations**:

```bash
# From web api (creates initial schema)
cd apps/web && bun run db:push

# From API (ensures schema is up to date)
cd apps/api && bun run db:push
```

5. **Start development servers**:

```bash
# Start both apps simultaneously
bun run dev

# Or start individually:
# Web api (http://localhost:5173)
cd apps/web && bun run dev

# API (http://localhost:3001)
cd apps/api && bun run dev
```

## Features Implemented

### Authentication System

- **Shared Authentication**: Both apps use the same session-based auth system
- **Lucia Integration**: Frontend handles login/signup with Lucia
- **Session Verification**: API validates sessions from the shared database
- **Middleware**: Auth middleware for protected API routes

### API Endpoints

**Authentication**:

- `GET /api/auth/verify` - Verify current session
- `GET /api/auth/session/:sessionId` - Get session info
- `DELETE /api/auth/logout` - Logout and delete session

**Users**:

- `GET /api/users` - Get all users (public)
- `GET /api/users/:id` - Get user by ID (public)
- `POST /api/users` - Create new user (public)
- `PUT /api/users/:id` - Update user (requires auth, own profile only)
- `DELETE /api/users/:id` - Delete user (requires auth, own profile only)
- `GET /api/users/me/profile` - Get current user profile (requires auth)

### MessagePack Communication

- Efficient binary serialization between frontend and backend
- Automatic encoding/decoding in API client
- Smaller payload sizes compared to JSON

## Testing the Setup

1. **Visit the demo page**: `http://localhost:5173/demo/api`
2. **Test API health**: Click "Test Health Check"
3. **Test user operations**: Click "Get Users" and "Create Test User"
4. **Check authentication**: Login via `/demo/lucia/login` and test protected endpoints

## Authentication Flow

### Frontend (Lucia)

1. User logs in via SvelteKit form
2. Lucia creates session in database
3. Session ID stored in secure cookie

### API Communication

1. Frontend sends session ID via header: `X-Session-Id: <sessionId>`
2. API validates session against database
3. Protected routes require valid session

### Session Sharing

Both apps use the same database schema, so sessions created in the frontend work seamlessly with the API.

## Development Workflow

### Frontend Development

```bash
cd apps/web
bun run dev          # Start dev server
bun run db:studio    # Open Drizzle Studio
bun run db:push      # Push schema changes
```

### Backend Development

```bash
cd apps/api
bun run dev          # Start dev server with hot reload
bun run db:studio    # Open Drizzle Studio
bun run build        # Build for production
```

### Monorepo Commands

```bash
bun run dev          # Start both apps
bun run build        # Build both apps
bun run lint         # Lint all packages
```

## API Client Usage

```typescript
import { apiClient } from '$lib/api';

// Health check
const health = await apiClient.healthCheck();

// Get users (public)
const users = await apiClient.getUsers();

// With authentication
const sessionId = 'user-session-id';
const currentUser = await apiClient.getCurrentUser(sessionId);

// Create user
const newUser = await apiClient.createUser({
	username: 'testuser',
	password: 'password123',
	age: 25,
});
```

## Security Features

- **Helmet**: Security headers on API
- **CORS**: Configured for development/production
- **Session Validation**: Automatic session expiry checking
- **Password Hashing**: Argon2 for secure password storage
- **Protected Routes**: Middleware-based authentication

## Production Deployment

1. **Build all apps**:

```bash
bun run build
```

2. **Environment variables**:

- Set production DATABASE_URL
- Configure CORS origins for production domains
- Set NODE_ENV=production

3. **Database**:

- Run migrations: `bun run db:migrate`
- Ensure connection pooling for production load

## Troubleshooting

### Common Issues

1. **Database connection errors**:

   - Check DATABASE_URL format
   - Ensure PostgreSQL is running
   - Verify credentials and database exists

2. **CORS errors**:

   - Check API CORS configuration in `apps/api/src/index.ts`
   - Ensure frontend URL matches CORS settings

3. **MessagePack errors**:

   - Verify `@msgpack/msgpack` is installed in web app
   - Check Content-Type headers in requests

4. **Session issues**:
   - Ensure both apps use same DATABASE_URL
   - Check session expiry in database
   - Verify session ID transmission (headers)

### Debug Commands

```bash
# Check database schema
cd apps/web && bun run db:studio

# Test API directly
curl http://localhost:3001/health

# Check logs
cd apps/api && bun run dev # Shows detailed logs
```

This setup provides a robust foundation for a full-stack application with efficient communication, shared authentication, and type-safe database operations.
