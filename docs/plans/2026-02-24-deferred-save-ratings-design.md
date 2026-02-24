# Deferred Save + Ratings Design

**Date:** 2026-02-24

## Summary

Restaurants are no longer saved to the DB on search. They are saved only when a logged-in user submits a rating. The search shows a Google Places preview card with a rating form. If the restaurant already exists in DB (rated by someone), `ListReviews` returns the current user's existing review for pre-population (edit mode).

---

## Decisions

| Concern | Decision |
|---|---|
| When to save restaurant | Only on rating submit — never on search |
| Auth gate | Must be logged in to search |
| Rating model | Stars (1–5) + optional comment + optional free-form tags |
| Tags storage | `text[]` Postgres array (no separate tag table) |
| Architecture | Single atomic `CreateReview` RPC creates-or-finds restaurant + saves review |
| Existing review | Pre-populated in edit mode; submit calls `UpdateReview` |
| Other users' reviews | Not shown (only current user's own review) |

---

## Proto Changes

### `packages/protos/reviews/v1/review.proto`

Change `tags` from `repeated TagProto` to `repeated string`:

```protobuf
message ReviewProto {
  string id = 1;
  string google_places_id = 2;
  string user_id = 3;
  string restaurant_id = 4;
  string comment = 5;
  double rating = 6;
  int64 created_at = 7;
  int64 updated_at = 8;
  repeated string tags = 9;
}
```

### `packages/protos/reviews/v1/reviews_service.proto`

Change `CreateReviewRequest` — takes Google Places data instead of DB restaurant_id (server creates-or-finds restaurant):

```protobuf
message CreateReviewRequest {
  string google_places_id = 1;
  string restaurant_name = 2;
  string restaurant_address = 3;
  string comment = 4;
  double rating = 5;
  repeated string tags = 6;
}

message CreateReviewResponse {
  ReviewProto review = 1;
  restaurants.v1.RestaurantProto restaurant = 2;
}
```

Add `google_places_id` filter to `ListReviewsRequest`:

```protobuf
message ListReviewsRequest {
  string restaurant_id = 1;
  string user_id = 2;
  string google_places_id = 5;  // new — filter by google places id
}
```

Remove `comment` and `rating` filter fields from `ListReviewsRequest` (unused).

Update `UpdateReviewRequest` to support updating tags:

```protobuf
message UpdateReviewRequest {
  string id = 1;
  string comment = 4;
  double rating = 5;
  repeated string tags = 6;
}
```

Add import: `import "restaurants/v1/restaurant.proto";` to reviews_service.proto.

---

## Backend: New Files

### `apps/api/src/internal/models/review_model.go`

```go
type Review struct {
    UUIDv7
    RestaurantID string         `gorm:"not null;index"`
    UserID       string         `gorm:"not null;index"`
    Comment      string
    Rating       float64        `gorm:"not null"`
    Tags         pq.StringArray `gorm:"type:text[]"`
    CreatedAt    time.Time
    UpdatedAt    time.Time
}
```

Unique constraint: `(restaurant_id, user_id)` — one review per user per restaurant.

### `apps/api/src/services/reviews_service.go`

Implements `ReviewsServiceHandler`:

**`CreateReview`:**
1. Extract user ID from session cookie (via `GetCurrentUser` helper or direct cache lookup)
2. In a DB transaction: `FirstOrCreate` a `Restaurant` record by `GoogleID`
3. Create a `Review` record `(restaurant_id, user_id, rating, comment, tags)`
4. Return `ReviewProto` + `RestaurantProto`

**`ListReviews`:**
- Filter by `google_places_id` (join restaurants table) + `user_id` from session
- Returns the current user's review for that place (0 or 1 result for the search use case)

**`UpdateReview`:**
- Updates `comment`, `rating`, `tags` for an existing review ID
- Validates the review belongs to the requesting user

**`GetReview`:** by review ID.

### `apps/api/src/main.go`

Register `ReviewsService` in `initializeServiceHandlers()`.

---

## Frontend Changes

### `apps/web/src/lib/client/client.ts`

Add `reviews` client:
```typescript
import { ReviewsService } from '$lib/client/generated/reviews/v1/reviews_service_pb';
const reviews = createClient(ReviewsService, transport);
export default { restaurants, users, googleMaps, auth, reviews };
```

### `apps/web/src/lib/ui/components/RestaurantSearchSv.svelte`

**Remove:** `createRestaurant` call entirely.

**Add auth gate:** If `!auth.isLoggedIn`, render a "Please log in to search restaurants" message instead of the search input.

**New state:**
```typescript
let existingReview = $state<ReviewProto | null>(null)
let submittedReview = $state<ReviewProto | null>(null)
let savedRestaurant = $state<RestaurantProto | null>(null)
let isCheckingReview = $state(false)
```

**On place selection:**
1. Set `selectedPlace`
2. Call `client.reviews.listReviews({ googlePlacesId: place.name })` (user_id from session)
3. If response has reviews[0]: set `existingReview` (edit mode)
4. Show `PlacePreviewCard` + `RatingForm`

**Render logic:**
- Not logged in → "Please log in to search"
- Logged in, no place selected → search input
- Place selected, loading → spinner
- Place selected → `PlacePreviewCard` (Google data) + `RatingForm` below it
- After submit → `RestaurantCard` (DB data from response) + `ReviewSummary`

### New: `apps/web/src/lib/ui/components/PlacePreviewCard.svelte`

Props: `{ place: Place }`

Shows read-only left panel (name + address from Google data, no edit controls, no DB created-date) + right panel toggle for Google details (same as current RestaurantCard right panel). No "Google details" button needed — can show details inline or collapsed.

### New: `apps/web/src/lib/ui/components/RatingForm.svelte`

Props: `{ googlePlacesId: string, restaurantName: string, restaurantAddress: string, existingReview?: ReviewProto, onSubmit: (review: ReviewProto, restaurant: RestaurantProto) => void }`

**UI elements:**
- Interactive star picker: 5 stars, click to set 1–5 rating (hover previews)
- Comment: `<textarea>` (optional)
- Tags: text input — press Enter or comma to add a chip; chips have an × to remove
- Submit button: "Save rating" (disabled until rating ≥ 1)
- Loading state during submit
- Error display on failure

**Logic:**
- If `existingReview` prop is set: pre-populate form + call `updateReview` on submit
- Else: call `createReview` on submit
- On success: call `onSubmit(review, restaurant)` callback

### New: `apps/web/src/lib/ui/components/ReviewSummary.svelte`

Props: `{ review: ReviewProto, onEdit: () => void }`

Shows: star display (read-only), comment, tag chips, "Edit" button that re-opens `RatingForm` in edit mode.

---

## Data Flow (Happy Path)

```
User logs in
  → search input appears
User types restaurant name
  → autocomplete suggestions
User selects a suggestion
  → getRestaurantDetails() → Place fetched
  → listReviews({ googlePlacesId, userId }) → [] (not in DB yet)
  → PlacePreviewCard + empty RatingForm shown
User fills in 4 stars + comment + tags, clicks "Save rating"
  → createReview({ googlePlacesId, name, address, rating, comment, tags })
  → Go: FirstOrCreate restaurant, create review
  → returns ReviewProto + RestaurantProto
  → RestaurantCard (DB data) + ReviewSummary shown
```

---

## Verification Checklist

- [ ] `bun run check` — 0 errors
- [ ] `cd apps/api && go build ./...` — compiles
- [ ] Search input hidden when not logged in
- [ ] Selecting a place does NOT create a DB record
- [ ] Submitting rating creates both restaurant + review in DB (check with DB query)
- [ ] Selecting same place as logged-in user shows pre-filled form (edit mode)
- [ ] Updating review via edit mode calls `UpdateReview`, not `CreateReview`
- [ ] No flowbite imports remain
