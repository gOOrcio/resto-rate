# Resto Rate API

Fastify-based API server with Drizzle ORM, msgpack communication, and shared authentication with the web app.

## Features

- **Fastify** - Fast and low overhead web framework
- **Drizzle ORM** - Type-safe database operations
- **MessagePack** - Efficient binary serialization for API communication
- **Shared Authentication** - Uses the same session system as the web app
- **CORS** - Configured for cross-origin requests
- **Security** - Helmet for security headers

## Setup

1. Install dependencies:

```bash
bun install
```

2. Set up environment variables:
   Create a `.env` file with:

```
DATABASE_URL="postgresql://username:password@localhost:5432/resto_rate"
PORT=3001
NODE_ENV=development
```

3. Run database migrations (if needed):

```bash
bun run db:push
```

4. Start the development server:

```bash
bun run dev
```

## API Endpoints

### Authentication

- `GET /api/auth/verify` - Verify current session
- `GET /api/auth/session/:sessionId` - Get session info
- `DELETE /api/auth/logout` - Logout and delete session

### Users

- `GET /api/users` - Get all users (public)
- `GET /api/users/:id` - Get user by ID (public)
- `POST /api/users` - Create new user (public)
- `PUT /api/users/:id` - Update user (requires auth, own profile only)
- `DELETE /api/users/:id` - Delete user (requires auth, own profile only)
- `GET /api/users/me/profile` - Get current user profile (requires auth)

### Health Check

- `GET /health` - Server health status

## Authentication

The API uses the same session-based authentication as the web app. Sessions are verified by:

1. **Authorization Header**: `Authorization: Bearer <sessionId>`
2. **Custom Header**: `X-Session-Id: <sessionId>`

The API shares the same database schema with the web app, so sessions created in the web app work seamlessly with the API.

## MessagePack Communication

All API responses can be returned in MessagePack format by setting the `Content-Type` header to `application/msgpack`. This provides more efficient serialization compared to JSON.

Example usage:

```javascript
// Request with msgpack response
fetch('/api/users', {
	headers: {
		'Content-Type': 'application/msgpack',
	},
});
```

## Development

- `bun run dev` - Start development server with hot reload
- `bun run build` - Build for production
- `bun run start` - Start production server
- `bun run db:push` - Push schema changes to database
- `bun run db:studio` - Open Drizzle Studio
