# Google OAuth Authentication Implementation Plan

## Overview

This document outlines the implementation plan for replacing the current username/password authentication system with Google OAuth authentication for the Resto Rate application. The goal is to provide a seamless, secure authentication experience using only Google OAuth.

## Current State Analysis

### Backend (Fastify API)
- **Current Auth System**: Username/password with Argon2 hashing
- **Session Management**: Database-based sessions with ULID tokens
- **Middleware**: `requireAuth` and `optionalAuth` middleware
- **Database Schema**: `user` table with `username`, `password_hash`, `age` fields
- **API Endpoints**: `/api/auth/login`, `/api/auth/register`, `/api/auth/verify`, `/api/auth/logout`

### Frontend (SvelteKit)
- **Current State**: No authentication state management
- **Navigation**: Basic navigation without auth checks
- **API Client**: MessagePack-based communication with backend
- **Stores**: Only theme store, no auth store

## Implementation Plan

### Phase 1: Database Schema Updates

#### 1.1 Update User Table Schema
```sql
-- Modify existing user table
ALTER TABLE "user" 
ADD COLUMN "google_id" text UNIQUE,
ADD COLUMN "email" text UNIQUE,
ADD COLUMN "name" text,
ADD COLUMN "is_admin" boolean DEFAULT false;

-- Make username and password_hash optional for OAuth users
ALTER TABLE "user" 
ALTER COLUMN "username" DROP NOT NULL,
ALTER COLUMN "password_hash" DROP NOT NULL;

-- Add indexes for performance
CREATE INDEX idx_user_google_id ON "user"("google_id");
CREATE INDEX idx_user_email ON "user"("email");
CREATE INDEX idx_user_is_admin ON "user"("is_admin");
```

#### 1.2 Update Drizzle Schema
```typescript
// packages/database/schema.ts
export const user = pgTable('user', {
	id: text('id').primaryKey(), // ULID
	googleId: text('google_id').unique(), // Google OAuth ID
	email: text('email').unique(), // Email from Google
	name: text('name'), // Full name from Google
	isAdmin: boolean('is_admin').default(false), // Future admin support
	username: text('username').unique(), // Optional for OAuth users
	passwordHash: text('password_hash'), // Optional for OAuth users
	age: integer('age'),
	createdAt: timestamp('created_at', { withTimezone: true, mode: 'date' }).defaultNow(),
	updatedAt: timestamp('updated_at', { withTimezone: true, mode: 'date' }).defaultNow(),
});
```

### Phase 2: Backend Implementation

#### 2.1 Environment Configuration
```typescript
// packages/config/env.ts
export interface AuthConfig {
	sessionSecret: string;
	sessionMaxAge: number; // 30 days in seconds
	googleClientId: string;
	googleClientSecret: string;
	googleRedirectUri: string;
}
```

#### 2.2 Google OAuth Service
```typescript
// apps/api/src/services/google-auth.service.ts
export interface GoogleUserInfo {
	id: string;
	email: string;
	verified_email: boolean;
	name: string;
	given_name: string;
	family_name: string;
	locale: string;
}

export async function exchangeCodeForTokens(code: string): Promise<{
	access_token: string;
	id_token: string;
}>;

export async function getGoogleUserInfo(accessToken: string): Promise<GoogleUserInfo>;

export async function createOrUpdateUserFromGoogle(googleUser: GoogleUserInfo): Promise<User>;
```

#### 2.3 Updated Auth Service
```typescript
// apps/api/src/services/auth.service.ts
export async function authenticateWithGoogle(code: string): Promise<AuthResponse> {
	// 1. Exchange authorization code for tokens
	// 2. Get user info from Google
	// 3. Create or update user in database (email + name only)
	// 4. Create session (30 days)
	// 5. Return auth response
}

export async function verifySession(sessionId: string): Promise<AuthResponse> {
	// Keep existing implementation
}
```

#### 2.4 New Auth Routes
```typescript
// apps/api/src/routes/auth.ts
export const authRoutes: FastifyPluginAsync = async (fastify) => {
	// Google OAuth callback
	fastify.get('/google/callback', async (request, reply) => {
		const { code } = request.query as { code: string };
		const result = await authService.authenticateWithGoogle(code);
		// Redirect to frontend with session token
	});

	// Get Google OAuth URL
	fastify.get('/google/url', async (request, reply) => {
		const authUrl = generateGoogleAuthUrl();
		return { authUrl };
	});

	// Keep existing verify and logout endpoints
};
```

### Phase 3: Frontend Implementation

#### 3.1 Authentication Store
```typescript
// apps/web/src/lib/stores/auth.ts
interface AuthState {
	user: User | null;
	sessionId: string | null;
	isLoading: boolean;
	isAuthenticated: boolean;
}

export const authStore = createAuthStore();
```

#### 3.2 Authentication Service
```typescript
// apps/web/src/lib/services/auth.service.ts
export class AuthService {
	async initiateGoogleLogin(): Promise<void>;
	async handleGoogleCallback(code: string): Promise<void>;
	async verifySession(): Promise<void>;
	async logout(): Promise<void>;
	async refreshSession(): Promise<void>;
}
```

#### 3.3 Protected Route Component
```svelte
<!-- apps/web/src/lib/components/ProtectedRoute.svelte -->
<script lang="ts">
	import { authStore } from '$lib/stores/auth';
	import { goto } from '$app/navigation';
	
	export let fallback: string = '/login';
	
	$: if (!$authStore.isAuthenticated && !$authStore.isLoading) {
		goto(fallback);
	}
</script>

{#if $authStore.isAuthenticated}
	<slot />
{:else if $authStore.isLoading}
	<div>Loading...</div>
{/if}
```

#### 3.4 Auth Guard Directive
```typescript
// apps/web/src/lib/directives/authGuard.ts
export function authGuard(node: HTMLElement, options: { 
	requireAuth?: boolean; 
	fallback?: string;
}) {
	// Hide/show elements based on auth state
}
```

### Phase 4: Google OAuth Setup

#### 4.1 Google Cloud Console Setup
1. Create new project or use existing
2. Enable Google+ API
3. Create OAuth 2.0 credentials
4. Configure authorized redirect URIs:
   - Development: `http://localhost:3001/api/auth/google/callback`
   - Production: `https://yourdomain.com/api/auth/google/callback`

#### 4.2 Environment Variables
```env
# Google OAuth Configuration
GOOGLE_CLIENT_ID=your-google-client-id
GOOGLE_CLIENT_SECRET=your-google-client-secret
GOOGLE_REDIRECT_URI=http://localhost:3001/api/auth/google/callback
```

### Phase 5: Frontend Pages & Components

#### 5.1 Updated Landing Page
```svelte
<!-- apps/web/src/routes/+page.svelte -->
<script lang="ts">
	import { authStore } from '$lib/stores/auth';
	import { Button } from '$lib/components/ui/button';
	
	async function handleGoogleLogin() {
		// Redirect to Google OAuth
	}
</script>

<div class="container mx-auto max-w-4xl p-6">
	<h1>Resto Rate</h1>
	
	{#if $authStore.isAuthenticated}
		<!-- Show authenticated user content -->
		<div>Welcome, {$authStore.user?.name}!</div>
	{:else}
		<!-- Show login options -->
		<Button on:click={handleGoogleLogin}>
			Sign in with Google
		</Button>
	{/if}
</div>
```

#### 5.2 Google OAuth Callback Page
```svelte
<!-- apps/web/src/routes/auth/callback/+page.svelte -->
<script lang="ts">
	import { onMount } from 'svelte';
	import { page } from '$app/stores';
	import { authService } from '$lib/services/auth.service';
	
	onMount(async () => {
		const code = $page.url.searchParams.get('code');
		if (code) {
			await authService.handleGoogleCallback(code);
		}
	});
</script>

<div>Processing authentication...</div>
```

#### 5.3 Protected Pages
```svelte
<!-- apps/web/src/routes/restaurants/+page.svelte -->
<script lang="ts">
	import ProtectedRoute from '$lib/components/ProtectedRoute.svelte';
</script>

<ProtectedRoute>
	<!-- Restaurant management content -->
</ProtectedRoute>
```

### Phase 6: Security & Best Practices

#### 6.1 Session Security
- Use secure, HTTP-only cookies for session storage
- Implement CSRF protection
- Add rate limiting for OAuth endpoints
- Validate Google tokens server-side

#### 6.2 Error Handling
- Handle OAuth errors gracefully with user-friendly messages
- Log authentication failures for monitoring
- Provide clear error states in UI
- Handle network failures and timeouts

#### 6.3 Data Privacy
- Only store email and name from Google
- Respect user privacy preferences
- Implement data deletion on account removal

### Phase 7: Migration Strategy

#### 7.1 Database Migration
```sql
-- Migration script to update existing users
-- Handle existing username/password users
-- Add Google OAuth fields
```

#### 7.2 Backward Compatibility
- Keep existing session system for transition period
- Allow both auth methods during migration
- Gradual deprecation of username/password

### Phase 8: Testing & Deployment

#### 8.1 Testing Strategy
- Unit tests for auth services
- Integration tests for OAuth flow
- E2E tests for complete auth journey
- Security testing for OAuth implementation
- Mobile responsiveness testing

#### 8.2 Deployment Checklist
- [ ] Google OAuth credentials configured
- [ ] Environment variables set
- [ ] Database migrations applied
- [ ] SSL certificates configured
- [ ] CORS settings updated
- [ ] Rate limiting configured

## Implementation Order

1. **Database Schema Updates** (Phase 1)
2. **Backend OAuth Service** (Phase 2)
3. **Frontend Auth Store** (Phase 3)
4. **Google OAuth Setup** (Phase 4)
5. **Frontend Pages** (Phase 5)
6. **Security Implementation** (Phase 6)
7. **Testing & Deployment** (Phase 8)

## Dependencies to Add

### Backend
```json
{
  "google-auth-library": "^9.0.0",
  "jsonwebtoken": "^9.0.0"
}
```

### Frontend
```json
{
  "@auth/core": "^0.18.0",
  "@auth/sveltekit": "^0.1.0"
}
```

## Clarified Requirements

✅ **User Data**: Store only email and name from Google  
✅ **Session Duration**: 30 days (current setting)  
✅ **Error Handling**: Comprehensive error handling with user-friendly messages  
✅ **Admin Users**: Prepare for future admin support (isAdmin field)  
✅ **Mobile Support**: Include mobile responsiveness testing  

## Success Criteria

- [ ] Users can sign in with Google OAuth only
- [ ] Landing page is accessible to all users
- [ ] Protected pages require authentication
- [ ] Session management is secure and reliable (30 days)
- [ ] Easy way to hide/show components based on auth state
- [ ] User ownership is properly tracked for entities
- [ ] Only email and name stored from Google
- [ ] Comprehensive error handling
- [ ] Mobile-responsive design
- [ ] No username/password authentication remains
- [ ] All existing functionality works with new auth system

## Timeline Estimate

- **Phase 1-2 (Backend)**: 2-3 days
- **Phase 3-5 (Frontend)**: 2-3 days
- **Phase 6-8 (Security & Testing)**: 1-2 days
- **Total**: 5-8 days

This plan provides a comprehensive roadmap for implementing Google OAuth authentication while maintaining security and user experience standards. 