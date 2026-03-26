# Phase 2 PR 1: Predefined Tags + Restaurant City/Country

**Date**: 2026-03-26
**Status**: Approved
**Phase**: 2a of MVP roadmap

---

## Goal

Replace free-form tag strings on reviews with a predefined tag list (seeded, served via `TagsService`). Add `city` and `country` fields to the `Restaurant` model. These are foundational data model changes required by Phase 3 (My Reviews filter) and Phase 4 (Wishlist filter).

---

## Scope

**In scope:**
- `Tag` model + `TagsService.ListTags` RPC (cached in Valkey)
- ~40 predefined tags seeded via always-run `SeedRequiredData` (production-safe, upsert by slug)
- `City` + `Country` on `Restaurant` model and proto
- `city` + `country` added to `CreateReviewRequest` (frontend sends from Place data)
- `TagPicker.svelte` — multi-select tag picker for use in forms
- `TagFilter.svelte` — `TagPicker` + AND/OR mode toggle for use in filter panels (Phase 3+)
- Replace free-form tag input in `RatingForm.svelte` with `TagPicker`

**Out of scope:**
- AND/OR filtering logic on the backend (Phase 3)
- Migrating existing free-form tag data on reviews (old slugs stay as-is)
- Tag management UI (admin only, post-MVP)

---

## Data Model

### Tag

New GORM model `apps/api/src/internal/models/tag_model.go`:

```go
type Tag struct {
    UUIDv7
    Slug      string `gorm:"uniqueIndex;not null"`
    Label     string `gorm:"not null"`
    Category  string `gorm:"not null;index"`
    CreatedAt time.Time `gorm:"autoCreateTime"`
}
```

### Restaurant (additions)

```go
City    string `gorm:"index"`
Country string `gorm:"index"`
```

GORM auto-migrates on startup. Both nullable (empty string for restaurants created before this change).

---

## Proto Changes

### New: `packages/protos/tags/v1/tag.proto`

```protobuf
syntax = "proto3";
package tags.v1;
option go_package = "api/src/generated/tags/v1";

message TagProto {
  string id = 1;
  string slug = 2;
  string label = 3;
  string category = 4;
}
```

### New: `packages/protos/tags/v1/tags_service.proto`

```protobuf
syntax = "proto3";
package tags.v1;
import "tags/v1/tag.proto";
option go_package = "api/src/generated/tags/v1";

service TagsService {
  rpc ListTags(ListTagsRequest) returns (ListTagsResponse);
}

message ListTagsRequest {}

message ListTagsResponse {
  repeated TagProto tags = 1;
}
```

### Modified: `packages/protos/restaurants/v1/restaurant.proto`

Add to `RestaurantProto`:
```protobuf
string city = 5;
string country = 6;
```

### Modified: `packages/protos/reviews/v1/reviews_service.proto`

Add to `CreateReviewRequest`:
```protobuf
string city = 7;
string country = 8;
```

---

## Backend

### TagsService (`apps/api/src/services/tags_service.go`)

```go
type TagsService struct {
    v1connect.UnimplementedTagsServiceHandler
    DB     *gorm.DB
    Valkey valkey.Client
}

func (s *TagsService) ListTags(ctx, req) (*connect.Response[tagsv1.ListTagsResponse], error) {
    // 1. Try cache: GET tags:all
    // 2. On miss: SELECT * FROM tags ORDER BY category, label
    // 3. Marshal to proto, SET tags:all (1h TTL)
    // 4. Return
}
```

Cache key: `tags:all`, TTL: 1 hour. Tags are static seed data — no cache invalidation needed.

### SeedRequiredData (`apps/api/src/internal/utils/database.go`)

New function, called unconditionally from `main.go` before `SeedDatabase`:

```go
func SeedRequiredData(db *gorm.DB) error {
    // Upsert all predefined tags by slug
    // INSERT INTO tags (...) ON CONFLICT (slug) DO UPDATE SET label=..., category=...
}
```

**Tag list (~40 tags, 6 categories):**

| Category | Slugs |
|----------|-------|
| Cuisine | `italian`, `japanese`, `mexican`, `chinese`, `indian`, `french`, `thai`, `american`, `mediterranean`, `korean` |
| Vibe | `romantic`, `casual`, `family-friendly`, `date-night`, `business-lunch`, `lively`, `quiet`, `trendy` |
| Price | `budget`, `mid-range`, `expensive`, `splurge` |
| Dietary | `vegan`, `vegetarian`, `gluten-free`, `halal`, `kosher`, `dairy-free` |
| Service | `fast-service`, `outdoor-seating`, `delivery`, `takeaway`, `reservations`, `dog-friendly` |
| Occasion | `birthday`, `anniversary`, `brunch`, `late-night` |

### reviews_service.go changes

`CreateReview` find-or-create restaurant: add `City` and `Country` from `req.Msg.City` / `req.Msg.Country`.

```go
restaurant := models.Restaurant{
    GoogleID: req.Msg.GooglePlacesId,
    Name:     req.Msg.RestaurantName,
    Address:  req.Msg.RestaurantAddress,
    City:     req.Msg.City,
    Country:  req.Msg.Country,
}
```

Note: `FirstOrCreate` won't update city/country on existing restaurants — acceptable for MVP (restaurants created before this change will have empty city/country until re-reviewed).

### main.go changes

```go
// After schema migration, before SeedDatabase:
if err := utils.SeedRequiredData(db); err != nil {
    slog.Error("Failed to seed required data", slog.Any("error", err))
    os.Exit(1)
}
```

Register `TagsService` in `initializeServiceHandlers`.

---

## Frontend

### New: `apps/web/src/lib/ui/components/TagPicker.svelte`

- Fetches tags from `client.tags.listTags({})` on mount
- Groups by category
- Multi-select: clicking a tag toggles it
- Props: `selected: string[]` (slugs), `onchange: (slugs: string[]) => void`
- No AND/OR toggle

### New: `apps/web/src/lib/ui/components/TagFilter.svelte`

- Wraps `TagPicker`
- Adds AND/OR mode toggle (default: OR)
- Props: `selected: string[]`, `mode: 'AND' | 'OR'`, `onchange: (slugs: string[], mode: 'AND' | 'OR') => void`
- Used in Phase 3 filter panels — built now so it's ready

### Modified: `apps/web/src/lib/ui/components/RatingForm.svelte`

- Replace free-form tag text input with `<TagPicker bind:selected={tags} />`
- `tags` sent as slug array — wire format unchanged

### Modified: `apps/web/src/lib/client/client.ts`

Add `tags` client:
```ts
import { TagsService } from '$lib/client/generated/tags/v1/tags_service_pb';
const tags = createClient(TagsService, transport);
export default { restaurants, users, googleMaps, auth, reviews, tags };
```

---

## Error Handling

| Scenario | Behaviour |
|----------|-----------|
| `ListTags` called with no tags seeded | Returns empty list (not an error) |
| Tag cache miss | Falls through to DB, re-populates cache |
| `city`/`country` empty on `CreateReviewRequest` | Stored as empty string — acceptable |
| Frontend `listTags` fails | TagPicker shows empty state with retry; form still submittable (tags optional) |

---

## What Is NOT In Scope

- AND/OR filter backend logic (Phase 3)
- Validating that submitted tag slugs exist in the tag table (post-MVP)
- Tag categories/labels editable via UI
- Backfilling city/country on existing restaurants
