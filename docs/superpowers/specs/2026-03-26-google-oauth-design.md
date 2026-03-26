# Google OAuth Design Spec

**Date**: 2026-03-26
**Status**: Approved
**Phase**: 1 of MVP roadmap

---

## Goal

Replace the placeholder username-only login with Google Sign-In using the Google Identity Services (GIS) SDK. Store `google_id`, `email`, `name` from the Google token. Username remains optional, falling back to email for display.

The design is explicitly structured to make Apple Sign-In easy to add later without proto or service changes.

---

## Approach

**Client-side GIS → ID token → Go AuthService RPC**

GIS JS library runs in the browser, handles the Google popup/One Tap flow, and returns a signed JWT credential. The frontend sends that token to `AuthService.Login` via the existing Connect-RPC call. Go verifies the JWT using Google's JWKS, extracts claims, upserts the user, and issues the existing Valkey session + HttpOnly cookie.

No redirect URIs needed. No new HTTP routes in Go. Plugs into the existing auth session architecture.

---

## Data Flow

```
Browser                          Go API (AuthService)
  │                                      │
  │  GIS JS loads (accounts.google.com/gsi/client)
  │  One Tap appears OR user clicks button
  │  Google popup → user picks account
  │  GIS returns credential (signed JWT)
  │                                      │
  ├──── Login RPC { provider: GOOGLE,    │
  │                 id_token: <jwt> } ──►│
  │                                      │  verifyIDToken() → ProviderClaims
  │                                      │  Upsert user by google_id
  │                                      │  Store session in Valkey (24h TTL)
  │◄──── LoginResponse (UserProto) ──────┤  Set-Cookie: session_token (HttpOnly)
  │                                      │
  │  auth.setUser(user)                  │
  │  UI updates (header shows name)      │
```

---

## Proto Changes

**File**: `packages/protos/auth/v1/auth_service.proto`

```protobuf
enum AuthProvider {
  AUTH_PROVIDER_UNSPECIFIED = 0;
  AUTH_PROVIDER_GOOGLE = 1;
  AUTH_PROVIDER_APPLE = 2;   // reserved — not implemented
}

message LoginRequest {
  AuthProvider provider = 1;
  string id_token = 2;
}
```

`LoginResponse` unchanged. `UserProto` unchanged (already has `google_id`, `email`, `name`, `username`).

---

## Backend

### New dependency
`google.golang.org/api/idtoken` — verifies Google JWT against Google's JWKS (handles key caching and rotation). No manual JWKS implementation needed.

### New env var
`GOOGLE_CLIENT_ID` — used as the expected audience in `idtoken.Validate()`. Fatal on startup if missing (same pattern as `GOOGLE_PLACES_API_KEY`).

### `auth_service.go` changes

```go
type ProviderClaims struct {
    ProviderID string  // "sub" from token — stable, never changes
    Email      string
    Name       string  // may be empty for Apple on repeat logins
}

func verifyIDToken(ctx, provider, token) (ProviderClaims, error)
// Currently: AUTH_PROVIDER_GOOGLE → idtoken.Validate()
// Future:    AUTH_PROVIDER_APPLE  → TODO (interface already defined)
```

**`Login` handler logic**:
1. Call `verifyIDToken()` — return `CodeUnauthenticated` on failure
2. Upsert user: `WHERE google_id = claims.ProviderID`
   - Found → update `email` + `name` (Google may change them)
   - Not found → create with `google_id`, `email`, `name`, `username = nil`
3. Issue Valkey session + Set-Cookie (existing logic, unchanged)

### `user_model.go` change
- `Username string` → `*string` (nullable)
- GORM auto-migrates the column on next startup

---

## Frontend

### New component: `SocialSignIn.svelte`
Replaces `LoginModal.svelte`.

- Loads GIS script dynamically on mount
- Initializes One Tap (`google.accounts.id.initialize`) — skipped if already logged in or if GIS detects WebView/unsupported context
- Renders the official Sign In With Google button (GIS handles mobile sizing via `data-width`)
- On credential callback: calls `clients.auth.login({ provider: AUTH_PROVIDER_GOOGLE, idToken: credential })` → `auth.setUser(user)`
- Has a clearly marked slot for a future Apple Sign-In button

### `+layout.svelte` changes
- After `getCurrentUser()` on mount: if not logged in, call `google.accounts.id.prompt()` to trigger One Tap
- One Tap only shown to unauthenticated users

### `Header.svelte` changes
- Logged out: "Sign in" button → opens `SocialSignIn` in a `<dialog>`
- Logged in: shows `user.username ?? user.email` + Logout
- Nav links (Reviews, Wishlist, Friends) only visible when logged in

### New env var
`VITE_GOOGLE_CLIENT_ID` in `apps/web/.env` — passed to GIS init.

---

## Error Handling

| Scenario | Backend | Frontend |
|----------|---------|----------|
| Invalid/expired token | `CodeUnauthenticated` | Toast: "Sign in failed, please try again" |
| One Tap dismissed | — | Silent; button remains visible |
| WebView / unsupported browser | — | GIS skips One Tap silently; button always rendered |
| First login (new user) | Create user, `username = nil` | Signed in immediately; no extra prompt |
| Returning user, changed Google email | Lookup by `google_id`, update email/name | Transparent |
| `GOOGLE_CLIENT_ID` missing (API) | Fatal on startup | — |
| `VITE_GOOGLE_CLIENT_ID` missing (web) | — | Console warning; button renders but GIS init fails |

---

## What Is NOT in Scope

- Apple Sign-In implementation (proto field reserved, Go interface defined, frontend slot reserved)
- Username prompt on first login (username is optional; profile editing is post-MVP)
- Profile photo storage
- Token refresh (Valkey TTL + re-login on expiry is sufficient for MVP)
- `/auth/google/callback` route (not needed — ID token flow has no redirect)
