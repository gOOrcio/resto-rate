# Resto Rate — MVP Roadmap

## Vision

Social restaurant rating app: find restaurants (Google Places), rate them (stars + comment + tags), build a wishlist, share with friends (mutual follow only).

Login via provider only — Google OAuth first, Apple later.

---

## Decisions Log

| # | Decision |
|---|----------|
| Tags | Predefined list (seeded to DB), not free-form |
| Tag filter mode | User-switchable: AND mode (all tags) OR mode (any tag) |
| Comment search | Within a specific list (rated/wishlist) of a specific user (self or friend) only; separate views per user per list |
| Geo filter | City + Country fields on Restaurant (extracted from Google Places address components); no lat/lng needed for MVP |
| Wishlist ↔ Review | Mutually exclusive: creating a review auto-removes the restaurant from the user's wishlist (atomic in DB transaction) |
| Friends | Mutual request/accept only — no one-way follow |
| Content visibility | Friends-only — reviews and wishlist not public |
| Review edit/delete | Both exposed in UI; no time limit on edits |
| Google OAuth data | Store: google_id, email, name. Skip: profile photo |
| Username | Optional; display falls back to Google email |

---

## Current State (as of Mar 2026)

### Done
- [x] Nx monorepo (bun), Docker infra (PostgreSQL + Valkey)
- [x] All 5 Connect-RPC services: `auth`, `users`, `restaurants`, `reviews`, `google_maps`
- [x] Session auth (placeholder — username only, to be replaced by Google OAuth)
- [x] Restaurant CRUD (unique by GoogleID + address)
- [x] Reviews: create/update/delete, 1–5 stars, comment, free-form tags (to be migrated to predefined), one review per user per restaurant
- [x] Google Places: autocomplete, details, text search
- [x] ShadCN Svelte UI migration (branch: `rewrite-to-shad-cn`)
- [x] UI: restaurant search, restaurant card, rating form, review summary, login modal
- [x] CI: SonarQube

### Known debt
- Auth is username-only; needs Google OAuth
- Tags are free-form strings; need migration to predefined tag table
- Restaurant has no `city`/`country` fields
- No routes beyond `/`
- `TagProto` in protos is unused (tags stored as `repeated string` on `ReviewProto`)
- Restaurant `Address` field has `json:"email"` tag (cosmetic bug)

---

## Implementation Phases

---

### Phase 1 — Google OAuth

**Why first**: Everything else requires authenticated users. Current auth is a placeholder.

#### Backend
- [ ] `LoginRequest` proto: add `google_id_token` field (replaces `username`)
- [ ] `AuthService.Login`: verify Google ID token (using Google tokeninfo endpoint or `google-auth-library`), extract `sub`/`email`/`name`, find-or-create User
- [ ] On first login: create user with `google_id`, `email`, `name`; `username` = nil
- [ ] On subsequent logins: find by `google_id`, issue session (existing flow)
- [ ] Remove username-only login path from `AuthService`
- [ ] Add auth middleware helper used by all mutations (returns `CodeUnauthenticated` if no valid session)

#### Frontend
- [ ] Replace `LoginModal` with Google Sign-In button (Google Identity Services JS SDK, `credential_response` → ID token)
- [ ] Send ID token to `auth.login()` RPC
- [ ] Remove username input from login flow
- [ ] Display user name (username ?? email) in Header

**Proto files to edit**: `auth/v1/auth_service.proto`
**Go files to edit**: `auth_service.go`, `user_model.go` (make username nullable)

---

### Phase 2 — Data Model & Schema Extensions

**Why second**: Multiple subsequent phases depend on these models existing.

#### Tags (predefined)
- [ ] New GORM model: `tag_model.go` — `Tag { ID, Slug, Label, Category?, CreatedAt }`
- [ ] New proto: `tags/v1/tag.proto` (reuse/replace existing `TagProto`), `tags_service.proto` with `ListTags` RPC
- [ ] New service: `TagsService` — `ListTags` (returns all tags, cached in Valkey)
- [ ] DB seed: large predefined tag list (categories: cuisine type, vibe, price, dietary, service, occasion, etc.)
- [ ] `ReviewProto.tags` stays as `repeated string` (slug values) — no schema change on Review
- [ ] Frontend: replace free-text tag input with multi-select tag picker from `ListTags`

#### Restaurant: city + country
- [ ] Add `City string` and `Country string` to `restaurant_model.go`
- [ ] Update `RestaurantProto`: add `city`, `country` fields
- [ ] Update `CreateRestaurant` / find-or-create in `ReviewsService`: extract city + country from Google Places `formattedAddress` or `addressComponents`
- [ ] Add `city` and `country` filter params to `ListRestaurants` (optional, for future use)

#### Wishlist
- [ ] New GORM model: `wishlist_item_model.go` — `WishlistItem { ID, UserID (indexed), RestaurantID (indexed), GooglePlacesID, CreatedAt }` — unique constraint on `(UserID, RestaurantID)`
- [ ] New proto: `wishlist/v1/wishlist_item.proto`, `wishlist_service.proto` with: `AddToWishlist`, `RemoveFromWishlist`, `ListWishlist` (with filter/sort params)
- [ ] New service: `WishlistService`
- [ ] Mutation in `ReviewsService.CreateReview`: within the existing DB transaction, delete any wishlist entry for `(userID, restaurantID)` before inserting the review

#### Friends
- [ ] New GORM model: `friendship_model.go` — `Friendship { ID, RequesterID (indexed), AddresseeID (indexed), Status (pending/accepted/declined), CreatedAt, UpdatedAt }` — unique constraint on `(RequesterID, AddresseeID)`
- [ ] New proto: `friends/v1/friendship.proto`, `friends_service.proto` with: `SendFriendRequest`, `AcceptFriendRequest`, `DeclineFriendRequest`, `Unfriend`, `ListFriends`, `ListPendingRequests`
- [ ] New service: `FriendsService`

**Proto files to create**: `tags/v1/`, `wishlist/v1/`, `friends/v1/`
**Go files to create**: `tag_model.go`, `wishlist_item_model.go`, `friendship_model.go`, `tags_service.go`, `wishlist_service.go`, `friends_service.go`
**Go files to edit**: `restaurant_model.go`, `reviews_service.go`, `main.go`

---

### Phase 3 — My Reviews Page (`/reviews`)

**Requires**: Phase 1 (auth), Phase 2 (tags, city/country)

#### Backend
- [ ] `ListReviewsRequest` proto: add filter + sort params:
  - `tag_slugs []string` + `tag_filter_mode enum { AND, OR }`
  - `min_rating`, `max_rating float64`
  - `comment_search string`
  - `city string`, `country string`
  - `sort_by enum { RATING_ASC, RATING_DESC, DATE_ASC, DATE_DESC }`
- [ ] `ListReviews` service: apply filters via GORM (tags → JSON contains, comment → `ILIKE`, city/country → JOIN restaurant table)
- [ ] Enforce: `user_id` in `ListReviewsRequest` must be the calling user OR a confirmed friend (check `Friendship` table)

#### Frontend
- [ ] Route: `/reviews`
- [ ] Review card component: restaurant name, stars, comment excerpt, tag chips, date, Edit/Delete buttons
- [ ] Filter sidebar / panel:
  - Tag multi-select with AND/OR mode toggle
  - Rating range slider (1–5)
  - Comment keyword search input
  - City / Country inputs
  - Sort dropdown
- [ ] Edit flow: opens `RatingForm` pre-filled with existing review
- [ ] Delete flow: confirmation dialog → `DeleteReview` RPC → remove card

---

### Phase 4 — Wishlist Page (`/wishlist`)

**Requires**: Phase 2 (wishlist model + service)

#### Backend
- [ ] `ListWishlistRequest` proto: add filter + sort params:
  - `city string`, `country string`
  - `tag_slugs []string` + `tag_filter_mode` (if wishlist items will support tags — see note below)
  - `sort_by enum { DATE_ASC, DATE_DESC, NAME_ASC, NAME_DESC }`
- [ ] `ListWishlist` service: join with Restaurant table for city/country filter

> **Note**: Wishlist items don't have tags or comments themselves. Tag filtering on wishlist would need to filter by tags on the *restaurant* (i.e. tags from other users' reviews of that restaurant) — which is complex. For MVP, filter wishlist by city/country + sort only. Add tag filter post-MVP.

#### Frontend
- [ ] Route: `/wishlist`
- [ ] On home search result: show "Add to Wishlist" button alongside "Review" button
- [ ] Wishlist card: restaurant name, address, city/country, date added, "Remove" button, "Review now" button (triggers review flow + removes from wishlist)
- [ ] Filter: city, country, sort

---

### Phase 5 — Friends (`/friends`, `/friends/[username]`)

**Requires**: Phase 2 (friends model + service), Phase 3 + 4 (for browsing friend content)

#### Backend
- [ ] `ListUsers` RPC: add `username_search string` filter (for friend lookup by username)
- [ ] `FriendsService` — all 6 RPCs
- [ ] `ListReviews` + `ListWishlist`: when `target_user_id` is provided, verify caller is a confirmed friend of that user

#### Frontend
- [ ] Route: `/friends` — list of confirmed friends + pending requests (incoming/outgoing)
  - Incoming requests: Accept / Decline buttons
  - Outgoing pending: Cancel button
  - Confirmed friends: "View profile" link + Unfriend button
  - Search box: find user by username → "Send Friend Request" button
- [ ] Route: `/friends/[username]` — friend's profile
  - Two tabs: **Reviews** / **Wishlist**
  - Same filter/sort UI as own reviews/wishlist pages (read-only — no edit/delete buttons)
  - Uses same `ListReviews` / `ListWishlist` RPCs with `target_user_id`

---

### Phase 6 — Polish (post-MVP, nice to have)

- [ ] Apple Sign-In
- [ ] Profile page (`/profile`) — edit username, view own stats
- [ ] Proper error + empty state pages
- [ ] Mobile responsive layout
- [ ] Rate limiting on API
- [ ] Pagination on all list views (currently missing on reviews/wishlist)
- [ ] Combined multi-friend view (aggregated reviews/wishlist from multiple friends)
- [ ] Wishlist tag support (tags on wishlist items, or aggregate from restaurant reviews)

---

## Route Map (MVP)

| Route | Auth required | Description |
|-------|--------------|-------------|
| `/` | No | Search + rate a restaurant (current) |
| `/reviews` | Yes | My rated restaurants (filter + sort) |
| `/wishlist` | Yes | My wishlist (filter + sort) |
| `/friends` | Yes | Friend list + requests |
| `/friends/[username]` | Yes (must be friends) | Friend's reviews + wishlist |

---

## Frontend Views

### Global / Navigation
- **Header** (all pages): logo, nav links (Home, My Reviews, Wishlist, Friends — auth-gated), user display name (username ?? email) + Logout. Links hidden when not logged in.
- **Auth gate**: unauthenticated users hitting a protected route → redirect to `/` with login prompt

---

### `/` — Home (existing, extended)
**Current**: Google Places search + rate flow.
**Add**:
- After search result appears: two action buttons — **"Rate it"** (opens RatingForm) and **"Add to Wishlist"**
- If user already reviewed this restaurant: show existing ReviewSummary with Edit option instead of RatingForm
- If restaurant is already wishlisted: show "Wishlisted ✓" with Remove option

---

### Login (no route — inline)
- **One Tap**: auto-appears as Google's floating prompt for signed-in Google users
- **Sign In With Google button**: always visible in Header when logged out; also shown in a modal on protected-route access
- On success: modal/prompt closes, header updates to show user name

---

### `/reviews` — My Reviews
**Layout**: filter panel (left sidebar on desktop, collapsible drawer on mobile) + scrollable card list (right)

**Review card**:
- Restaurant name + address
- Star rating (visual, 1–5)
- Comment excerpt (truncated, expand on click)
- Tag chips
- Date reviewed
- **Edit** button → opens RatingForm pre-filled
- **Delete** button → inline confirmation ("Delete this review?") → gone

**Filter panel**:
- Tag multi-select (searchable list of predefined tags) with AND / OR toggle
- Rating range: min/max star selector
- Keyword search in comments (debounced input)
- City + Country text filters
- Sort: Rating ↑, Rating ↓, Newest, Oldest

**Empty state**: "You haven't rated any restaurants yet. [Search for one →]"

---

### `/wishlist` — My Wishlist
**Layout**: same filter+list pattern as `/reviews`

**Wishlist card**:
- Restaurant name + address + city, country
- Date added
- **"Review now"** button → opens RatingForm; on submit, card disappears from wishlist
- **Remove** button → inline confirmation

**Filter panel**:
- City + Country inputs
- Sort: Newest, Oldest, Name A–Z, Name Z–A

**Empty state**: "Your wishlist is empty. Find a restaurant to save for later. [Search →]"

---

### `/friends` — Friends
**Layout**: two sections stacked

**Section 1 — Pending requests**:
- Incoming: avatar/name + "Accept" / "Decline" buttons
- Outgoing: avatar/name + "Cancel request" button
- Hidden if no pending requests

**Section 2 — My friends**:
- Search box: find user by username → if found, show name + "Send Friend Request" button
- Friend list: each entry shows display name, username, **"View profile"** link, **"Unfriend"** button (with confirmation)

**Empty state**: "No friends yet. Search for someone by username."

---

### `/friends/[username]` — Friend Profile
**Header**: display name + username + "Unfriend" button

**Tabs**: Reviews | Wishlist

**Reviews tab** (read-only):
- Same filter panel as `/reviews` (no Edit/Delete buttons on cards)
- If friend has no reviews: "No reviews yet."

**Wishlist tab** (read-only):
- Same filter panel as `/wishlist`
- If friend has no wishlist items: "Empty wishlist."

**Access control**: if not friends with this user → show "You need to be friends to view this profile."

---

### Shared Components (new/updated)
| Component | Description |
|-----------|-------------|
| `SocialSignIn.svelte` | Google button + One Tap init; slot for future Apple button |
| `ReviewCard.svelte` | Reusable review card for `/reviews` + `/friends/[username]` tabs |
| `WishlistCard.svelte` | Reusable wishlist card |
| `FilterPanel.svelte` | Reusable filter sidebar; accepts config for which filters to show |
| `TagPicker.svelte` | Multi-select from predefined tag list with AND/OR mode toggle |
| `ConfirmButton.svelte` | Inline "are you sure?" pattern for Delete/Remove/Unfriend |
| `EmptyState.svelte` | Reusable empty state with icon + message + optional CTA |

---

## Proto Changes Summary

| Proto | Change |
|-------|--------|
| `auth/v1/auth_service.proto` | `LoginRequest`: replace `username` with `google_id_token` |
| `restaurants/v1/restaurant.proto` | Add `city`, `country` fields |
| `reviews/v1/reviews_service.proto` | `ListReviewsRequest`: add filter/sort params; add `target_user_id` |
| `tags/v1/` | New — `Tag` message + `TagsService.ListTags` |
| `wishlist/v1/` | New — `WishlistItem` message + `WishlistService` (Add, Remove, List) |
| `friends/v1/` | New — `Friendship` message + `FriendsService` (6 RPCs) |
| `users/v1/users_service.proto` | `ListUsersRequest`: add `username_search` |
