# Phase 2 PR1: Predefined Tags + Restaurant City/Country — Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Replace free-form tag input on reviews with a predefined tag list served by a new `TagsService`, and add `city`/`country` fields to `Restaurant`.

**Architecture:** New `Tag` GORM model seeded via an always-running `SeedRequiredData` (upsert by slug, production-safe). `TagsService.ListTags` serves the tag list cached in Valkey (`tags:all`, 1h TTL). `Restaurant` gains `City`/`Country` fields populated from `CreateReviewRequest`. Frontend gets two new components: `TagPicker` (multi-select, no toggle) and `TagFilter` (TagPicker + AND/OR toggle for Phase 3 filter panels).

**Tech Stack:** Go + Connect-RPC, GORM, Valkey, Protocol Buffers (Buf CLI codegen), SvelteKit 5 (Svelte runes), Tailwind CSS, shadcn-svelte, lucide-svelte.

---

## File Map

**Create:**
- `packages/protos/tags/v1/tag.proto` — `TagProto` message
- `packages/protos/tags/v1/tags_service.proto` — `TagsService` with `ListTags` RPC
- `apps/api/src/internal/models/tag_model.go` — `Tag` GORM model + `ToProto()`
- `apps/api/src/services/tags_service.go` — `TagsService` Connect-RPC handler
- `apps/api/src/test/tags_service_test.go` — `TestListTags_EmptyDB` test
- `apps/web/src/lib/ui/components/TagPicker.svelte` — multi-select tag picker
- `apps/web/src/lib/ui/components/TagFilter.svelte` — TagPicker + AND/OR toggle

**Modify:**
- `packages/protos/restaurants/v1/restaurant.proto` — add `city=5`, `country=6`
- `packages/protos/reviews/v1/reviews_service.proto` — add `city=7`, `country=8` to `CreateReviewRequest`
- `apps/api/src/internal/models/restaurant_model.go` — add `City`, `Country` fields + update `ToProto()`
- `apps/api/src/internal/utils/database.go` — add `SeedRequiredData()` + `AutoMigrate(&models.Tag{})`
- `apps/api/src/services/reviews_service.go` — add `City`/`Country` to `FirstOrCreate` attrs
- `apps/api/src/main.go` — call `SeedRequiredData`, register `TagsService`, add to gRPC reflection
- `apps/web/src/lib/client/client.ts` — add `tags` client
- `apps/web/src/lib/ui/components/RatingForm.svelte` — replace free-form tag input with `TagPicker`

---

## Task 1: Tag proto + codegen

**Files:**
- Create: `packages/protos/tags/v1/tag.proto`
- Create: `packages/protos/tags/v1/tags_service.proto`
- Modify: `packages/protos/restaurants/v1/restaurant.proto`
- Modify: `packages/protos/reviews/v1/reviews_service.proto`

- [ ] **Step 1: Create `packages/protos/tags/v1/tag.proto`**

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

- [ ] **Step 2: Create `packages/protos/tags/v1/tags_service.proto`**

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

- [ ] **Step 3: Add `city` and `country` to `packages/protos/restaurants/v1/restaurant.proto`**

The current proto has fields 1–4 and 6–7 (field 5 is unused). Add:

```protobuf
message RestaurantProto {
  string id = 1;
  string google_places_id = 2;
  string name = 3;
  string address = 4;
  string city = 5;
  string country = 6;
  int64 created_at = 7;
  int64 updated_at = 8;
}
```

> Note: `created_at` moves from field 6→7 and `updated_at` from 7→8. This is a breaking wire-format change but fine — no backwards compatibility needed.

- [ ] **Step 4: Add `city` and `country` to `CreateReviewRequest` in `packages/protos/reviews/v1/reviews_service.proto`**

```protobuf
message CreateReviewRequest {
  string google_places_id = 1;
  string restaurant_name = 2;
  string restaurant_address = 3;
  string comment = 4;
  double rating = 5;
  repeated string tags = 6;
  string city = 7;
  string country = 8;
}
```

- [ ] **Step 5: Regenerate Go protos**

```bash
cd /home/gooral/Projects/resto-rate
nx run protos:generate:api
```

Expected: no errors, new files in `apps/api/src/generated/tags/v1/`.

- [ ] **Step 6: Verify Go compiles with proto changes**

```bash
cd apps/api && go build ./...
```

Expected: may fail on missing `tags_service.go` — that's fine (undefined reference errors only, not proto errors). If there are proto-related errors, fix the `.proto` files.

- [ ] **Step 7: Regenerate TypeScript protos**

```bash
cd /home/gooral/Projects/resto-rate
nx run protos:generate:web
```

Expected: new TS files in `apps/web/src/lib/client/generated/tags/v1/`.

- [ ] **Step 8: Force-add generated TS tag files (they're gitignored)**

```bash
git add -f apps/web/src/lib/client/generated/tags/
```

- [ ] **Step 9: Commit**

```bash
git add packages/protos/tags/ packages/protos/restaurants/v1/restaurant.proto packages/protos/reviews/v1/reviews_service.proto apps/api/src/generated/ apps/web/src/lib/client/generated/tags/
git commit -m "feat: add tags proto + city/country to restaurant and review protos"
```

---

## Task 2: Tag model + SeedRequiredData

**Files:**
- Create: `apps/api/src/internal/models/tag_model.go`
- Modify: `apps/api/src/internal/utils/database.go`

- [ ] **Step 1: Write the failing test**

The existing test suite has no DB dependency (no SQLite in go.mod). Expose the tag list as `utils.RequiredTags` so it can be verified without a DB connection.

Create `apps/api/src/test/tags_service_test.go`:

```go
package test

import (
	"api/src/internal/utils"
	"testing"
)

func TestRequiredTags_Count(t *testing.T) {
	if got := len(utils.RequiredTags); got != 40 {
		t.Fatalf("expected 40 required tags, got %d", got)
	}
}

func TestRequiredTags_Slugs_Unique(t *testing.T) {
	seen := make(map[string]bool)
	for _, tag := range utils.RequiredTags {
		if seen[tag.Slug] {
			t.Fatalf("duplicate slug: %s", tag.Slug)
		}
		seen[tag.Slug] = true
		if tag.Slug == "" {
			t.Fatal("tag has empty slug")
		}
		if tag.Label == "" {
			t.Fatal("tag has empty label")
		}
		if tag.Category == "" {
			t.Fatal("tag has empty category")
		}
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

```bash
cd apps/api && go test ./src/test/... -run TestRequiredTags -v
```

Expected: FAIL — `utils.RequiredTags` undefined.

- [ ] **Step 3: Create `apps/api/src/internal/models/tag_model.go`**

```go
package models

import (
	tagsv1 "api/src/generated/tags/v1"
	"time"

	"gorm.io/gorm"
)

type Tag struct {
	UUIDv7
	Slug      string    `gorm:"uniqueIndex;not null"`
	Label     string    `gorm:"not null"`
	Category  string    `gorm:"not null;index"`
	CreatedAt time.Time `gorm:"autoCreateTime"`
}

func (t *Tag) BeforeCreate(tx *gorm.DB) (err error) {
	return t.UUIDv7.BeforeCreate(tx)
}

func (t *Tag) ToProto() *tagsv1.TagProto {
	return &tagsv1.TagProto{
		Id:       t.ID,
		Slug:     t.Slug,
		Label:    t.Label,
		Category: t.Category,
	}
}
```

- [ ] **Step 4: Add `SeedRequiredData` and `RequiredTags` to `apps/api/src/internal/utils/database.go`**

Add the `gorm.io/gorm/clause` import to the existing imports block:

```go
import (
	"api/src/internal/models"
	"log/slog"
	"os"
	"strings"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)
```

Add these exported var and function before the existing `CreateSchema`:

```go
// RequiredTags is the canonical list of predefined tags seeded into the DB.
// Exported so tests can verify the list without a DB connection.
var RequiredTags = []models.Tag{
	// Cuisine
	{Slug: "italian", Label: "Italian", Category: "Cuisine"},
	{Slug: "japanese", Label: "Japanese", Category: "Cuisine"},
	{Slug: "mexican", Label: "Mexican", Category: "Cuisine"},
	{Slug: "chinese", Label: "Chinese", Category: "Cuisine"},
	{Slug: "indian", Label: "Indian", Category: "Cuisine"},
	{Slug: "french", Label: "French", Category: "Cuisine"},
	{Slug: "thai", Label: "Thai", Category: "Cuisine"},
	{Slug: "american", Label: "American", Category: "Cuisine"},
	{Slug: "mediterranean", Label: "Mediterranean", Category: "Cuisine"},
	{Slug: "korean", Label: "Korean", Category: "Cuisine"},
	// Vibe
	{Slug: "romantic", Label: "Romantic", Category: "Vibe"},
	{Slug: "casual", Label: "Casual", Category: "Vibe"},
	{Slug: "family-friendly", Label: "Family Friendly", Category: "Vibe"},
	{Slug: "date-night", Label: "Date Night", Category: "Vibe"},
	{Slug: "business-lunch", Label: "Business Lunch", Category: "Vibe"},
	{Slug: "lively", Label: "Lively", Category: "Vibe"},
	{Slug: "quiet", Label: "Quiet", Category: "Vibe"},
	{Slug: "trendy", Label: "Trendy", Category: "Vibe"},
	// Price
	{Slug: "budget", Label: "Budget", Category: "Price"},
	{Slug: "mid-range", Label: "Mid-Range", Category: "Price"},
	{Slug: "expensive", Label: "Expensive", Category: "Price"},
	{Slug: "splurge", Label: "Splurge", Category: "Price"},
	// Dietary
	{Slug: "vegan", Label: "Vegan", Category: "Dietary"},
	{Slug: "vegetarian", Label: "Vegetarian", Category: "Dietary"},
	{Slug: "gluten-free", Label: "Gluten-Free", Category: "Dietary"},
	{Slug: "halal", Label: "Halal", Category: "Dietary"},
	{Slug: "kosher", Label: "Kosher", Category: "Dietary"},
	{Slug: "dairy-free", Label: "Dairy-Free", Category: "Dietary"},
	// Service
	{Slug: "fast-service", Label: "Fast Service", Category: "Service"},
	{Slug: "outdoor-seating", Label: "Outdoor Seating", Category: "Service"},
	{Slug: "delivery", Label: "Delivery", Category: "Service"},
	{Slug: "takeaway", Label: "Takeaway", Category: "Service"},
	{Slug: "reservations", Label: "Reservations", Category: "Service"},
	{Slug: "dog-friendly", Label: "Dog Friendly", Category: "Service"},
	// Occasion
	{Slug: "birthday", Label: "Birthday", Category: "Occasion"},
	{Slug: "anniversary", Label: "Anniversary", Category: "Occasion"},
	{Slug: "brunch", Label: "Brunch", Category: "Occasion"},
	{Slug: "late-night", Label: "Late Night", Category: "Occasion"},
}

// SeedRequiredData seeds production-required data unconditionally using upsert.
// Safe to call on every startup — idempotent by slug.
func SeedRequiredData(db *gorm.DB) error {
	slog.Info("Seeding required data (tags)...")

	// Copy the list so GORM's ID-setting side-effect doesn't mutate RequiredTags
	tags := make([]models.Tag, len(RequiredTags))
	copy(tags, RequiredTags)

	result := db.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "slug"}},
		DoUpdates: clause.AssignmentColumns([]string{"label", "category"}),
	}).Create(&tags)
	if result.Error != nil {
		return result.Error
	}

	slog.Info("Required data seeded successfully", slog.Int64("tags", int64(len(tags))))
	return nil
}
```

- [ ] **Step 5: Add `Tag` to `CreateSchema` in `database.go`**

In the existing `CreateSchema` function, add `Tag` migration after `Review`:

```go
if err := db.AutoMigrate(&models.Tag{}); err != nil {
    return err
}
```

- [ ] **Step 6: Run the test to verify it passes**

```bash
cd apps/api && go test ./src/test/... -run TestSeedRequiredData_IdempotentTags -v
```

Expected: PASS.

> Note: This test uses SQLite in-memory (`gorm.io/driver/sqlite`). If sqlite is not in go.mod, check with `grep sqlite go.mod`. If missing, add it: `go get gorm.io/driver/sqlite` — or look at existing tests to see what driver they use for testing (they may use a real Postgres test DB instead, in which case follow that pattern).

- [ ] **Step 7: Verify Go builds**

```bash
cd apps/api && go build ./...
```

Expected: may still fail on missing `tags_service.go` — that's fine for now.

- [ ] **Step 8: Commit**

```bash
git add apps/api/src/internal/models/tag_model.go apps/api/src/internal/utils/database.go apps/api/src/test/tags_service_test.go
git commit -m "feat: add Tag model and SeedRequiredData for predefined tags"
```

---

## Task 3: TagsService (Go backend)

**Files:**
- Create: `apps/api/src/services/tags_service.go`

- [ ] **Step 1: Add test for `TagsService` with nil DB returns error**

Add to `apps/api/src/test/tags_service_test.go`:

```go
func TestListTags_NilDB_ReturnsError(t *testing.T) {
	// Verify that TagsService with nil DB and nil Valkey returns an error
	// (rather than panicking). Tests the happy-path logic exists — real
	// DB integration is verified by the running app.
	svc := &services.TagsService{} // nil DB, nil Valkey

	req := connect.NewRequest(&tagsv1.ListTagsRequest{})
	_, err := svc.ListTags(context.Background(), req)
	if err == nil {
		t.Fatal("expected error from nil DB, got nil")
	}
}
```

Update the imports at top of `tags_service_test.go`:

```go
import (
	"api/src/internal/utils"
	"api/src/services"
	tagsv1 "api/src/generated/tags/v1"
	"context"
	"testing"

	"connectrpc.com/connect"
)
```

- [ ] **Step 2: Run test to verify it fails**

```bash
cd apps/api && go test ./src/test/... -run TestListTags_NilDB_ReturnsError -v
```

Expected: FAIL — `services.TagsService` undefined.

- [ ] **Step 3: Create `apps/api/src/services/tags_service.go`**

```go
package services

import (
	tagsv1 "api/src/generated/tags/v1"
	"api/src/generated/tags/v1/v1connect"
	"api/src/internal/cache"
	"api/src/internal/models"
	"context"
	"time"

	"connectrpc.com/connect"
	"github.com/valkey-io/valkey-go"
	"gorm.io/gorm"
)

const tagsCacheKey = "tags:all"
const tagsCacheTTL = time.Hour

type TagsService struct {
	v1connect.UnimplementedTagsServiceHandler
	DB     *gorm.DB
	Valkey valkey.Client
}

func NewTagsService(db *gorm.DB, kv valkey.Client) *TagsService {
	return &TagsService{DB: db, Valkey: kv}
}

func (s *TagsService) ListTags(
	ctx context.Context,
	_ *connect.Request[tagsv1.ListTagsRequest],
) (*connect.Response[tagsv1.ListTagsResponse], error) {
	// Try cache first (skip if Valkey is nil, e.g. in tests)
	if s.Valkey != nil {
		pc := cache.NewProtoCache(s.Valkey, tagsCacheTTL, "")
		cached := &tagsv1.ListTagsResponse{}
		if ok, _ := pc.Get(ctx, tagsCacheKey, cached); ok {
			return connect.NewResponse(cached), nil
		}
	}

	var tags []models.Tag
	if err := s.DB.WithContext(ctx).Order("category, label").Find(&tags).Error; err != nil {
		return nil, err
	}

	protos := make([]*tagsv1.TagProto, len(tags))
	for i, t := range tags {
		protos[i] = t.ToProto()
	}
	resp := &tagsv1.ListTagsResponse{Tags: protos}

	// Populate cache
	if s.Valkey != nil {
		pc := cache.NewProtoCache(s.Valkey, tagsCacheTTL, "")
		pc.Set(ctx, tagsCacheKey, resp)
	}

	return connect.NewResponse(resp), nil
}
```

- [ ] **Step 4: Run tests to verify they pass**

```bash
cd apps/api && go test ./src/test/... -run TestListTags -v
cd apps/api && go test ./src/test/... -run TestRequiredTags -v
```

Expected: both PASS.

- [ ] **Step 5: Verify Go builds**

```bash
cd apps/api && go build ./...
```

Expected: PASS (tags_service.go compiles, main.go not yet wired up so no registration error).

- [ ] **Step 6: Commit**

```bash
git add apps/api/src/services/tags_service.go apps/api/src/test/tags_service_test.go
git commit -m "feat: add TagsService with ListTags and Valkey caching"
```

---

## Task 4: Wire TagsService + SeedRequiredData into main.go

**Files:**
- Modify: `apps/api/src/main.go`

- [ ] **Step 1: Update `main.go` — call `SeedRequiredData` before `SeedDatabase`**

In `main()`, between `CreateSchema` and `SeedDatabase`:

```go
err = utils.SeedRequiredData(db)
if err != nil {
    slog.Error("Failed to seed required data", slog.Any("error", err))
    os.Exit(1)
}
```

The full startup sequence in `main()` becomes:

```go
err := utils.CreateSchema(db)
if err != nil {
    slog.Error("Failed to create database schema", slog.Any("error", err))
    os.Exit(1)
}

err = utils.SeedRequiredData(db)
if err != nil {
    slog.Error("Failed to seed required data", slog.Any("error", err))
    os.Exit(1)
}

err = utils.SeedDatabase(db)
if err != nil {
    slog.Error("Failed to seed database", slog.Any("error", err))
    os.Exit(1)
}
```

- [ ] **Step 2: Register `TagsService` in `initializeServiceHandlers`**

Add this import at the top of `main.go`:

```go
tagsv1connect "api/src/generated/tags/v1/v1connect"
```

Add service registration to `initializeServiceHandlers` return slice:

```go
func() ServiceRegistration {
    svc := services.NewTagsService(db, valkeyClient)
    path, handler := tagsv1connect.NewTagsServiceHandler(svc, connect.WithInterceptors(prometheusInterceptor))
    return ServiceRegistration{Path: path, Handler: handler}
}(),
```

- [ ] **Step 3: Add `TagsService` to gRPC reflection in `optionallySetupGRPCReflection`**

```go
reflector := grpcreflect.NewStaticReflector(
    usersv1connect.UsersServiceName,
    restaurantsv1connect.RestaurantsServiceName,
    googlemapsv1connect.GoogleMapsServiceName,
    authv1connect.AuthServiceName,
    reviewsv1connect.ReviewsServiceName,
    tagsv1connect.TagsServiceName,
)
```

- [ ] **Step 4: Verify Go builds**

```bash
cd apps/api && go build ./...
```

Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git add apps/api/src/main.go
git commit -m "feat: wire TagsService and SeedRequiredData into main.go"
```

---

## Task 5: Restaurant city/country — model + reviews_service

**Files:**
- Modify: `apps/api/src/internal/models/restaurant_model.go`
- Modify: `apps/api/src/services/reviews_service.go`

- [ ] **Step 1: Update `restaurant_model.go`**

Add `City` and `Country` fields and update `ToProto()`. The field numbers in `RestaurantProto` changed in Task 1 (created_at is now field 7, updated_at is field 8).

Replace the file content:

```go
package models

import (
	restaurantpb "api/src/generated/restaurants/v1"
	"time"

	"gorm.io/gorm"
)

type Restaurant struct {
	UUIDv7
	GoogleID  string    `gorm:"uniqueIndex" json:"googleId"`
	Address   string    `gorm:"uniqueIndex" json:"email"`
	Name      string    `gorm:"not null" json:"name"`
	City      string    `gorm:"index"`
	Country   string    `gorm:"index"`
	CreatedAt time.Time `gorm:"autoCreateTime" json:"createdAt"`
	UpdatedAt time.Time `gorm:"autoUpdateTime" json:"updatedAt"`
}

func (r *Restaurant) BeforeCreate(tx *gorm.DB) (err error) {
	if err = r.UUIDv7.BeforeCreate(tx); err != nil {
		return err
	}
	return nil
}

func (r *Restaurant) ToProto() *restaurantpb.RestaurantProto {
	return &restaurantpb.RestaurantProto{
		Id:             r.ID,
		GooglePlacesId: r.GoogleID,
		Address:        r.Address,
		Name:           r.Name,
		City:           r.City,
		Country:        r.Country,
		CreatedAt:      r.CreatedAt.Unix(),
		UpdatedAt:      r.UpdatedAt.Unix(),
	}
}
```

- [ ] **Step 2: Update `reviews_service.go` — populate City/Country in FirstOrCreate**

In `CreateReview`, change the `Attrs` call to include `City` and `Country`:

```go
result := tx.Where(models.Restaurant{GoogleID: req.Msg.GooglePlacesId}).
    Attrs(models.Restaurant{
        Name:    req.Msg.RestaurantName,
        Address: req.Msg.RestaurantAddress,
        City:    req.Msg.City,
        Country: req.Msg.Country,
    }).
    FirstOrCreate(&restaurant)
```

- [ ] **Step 3: Verify Go builds**

```bash
cd apps/api && go build ./...
```

Expected: PASS.

- [ ] **Step 4: Run all Go tests**

```bash
cd apps/api && go test ./src/test/... -v
```

Expected: all PASS.

- [ ] **Step 5: Commit**

```bash
git add apps/api/src/internal/models/restaurant_model.go apps/api/src/services/reviews_service.go
git commit -m "feat: add city/country to Restaurant model and CreateReview"
```

---

## Task 6: Frontend — TagPicker.svelte

**Files:**
- Create: `apps/web/src/lib/ui/components/TagPicker.svelte`

- [ ] **Step 1: Create `apps/web/src/lib/ui/components/TagPicker.svelte`**

```svelte
<script lang="ts">
	import client from '$lib/client/client';
	import type { TagProto } from '$lib/client/generated/tags/v1/tag_pb';

	const { selected = $bindable([]), onchange } = $props<{
		selected?: string[];
		onchange?: (slugs: string[]) => void;
	}>();

	let tags = $state<TagProto[]>([]);
	let loading = $state(true);
	let loadError = $state(false);

	async function loadTags() {
		loading = true;
		loadError = false;
		try {
			const res = await client.tags.listTags({});
			tags = res.tags;
		} catch {
			loadError = true;
		} finally {
			loading = false;
		}
	}

	$effect(() => {
		loadTags();
	});

	function toggleTag(slug: string) {
		const next = selected.includes(slug)
			? selected.filter((s) => s !== slug)
			: [...selected, slug];
		selected = next;
		onchange?.(next);
	}

	// Group tags by category, preserving server order
	const grouped = $derived(
		tags.reduce(
			(acc, tag) => {
				if (!acc[tag.category]) acc[tag.category] = [];
				acc[tag.category].push(tag);
				return acc;
			},
			{} as Record<string, TagProto[]>
		)
	);
</script>

{#if loading}
	<p class="text-sm text-gray-400">Loading tags…</p>
{:else if loadError}
	<div class="flex items-center gap-2 text-sm text-red-500">
		<span>Failed to load tags.</span>
		<button type="button" onclick={loadTags} class="underline hover:no-underline">Retry</button>
	</div>
{:else if tags.length === 0}
	<p class="text-sm text-gray-400">No tags available.</p>
{:else}
	<div class="flex flex-col gap-3">
		{#each Object.entries(grouped) as [category, categoryTags]}
			<div>
				<p class="mb-1 text-xs font-semibold uppercase tracking-wide text-gray-400">{category}</p>
				<div class="flex flex-wrap gap-1.5">
					{#each categoryTags as tag}
						<button
							type="button"
							onclick={() => toggleTag(tag.slug)}
							class="rounded-full px-2.5 py-0.5 text-xs font-medium transition-colors
								{selected.includes(tag.slug)
								? 'bg-blue-600 text-white'
								: 'bg-gray-100 text-gray-600 hover:bg-gray-200'}"
						>
							{tag.label}
						</button>
					{/each}
				</div>
			</div>
		{/each}
	</div>
{/if}
```

- [ ] **Step 2: Run svelte-check to verify no type errors**

```bash
cd apps/web && bun run check
```

Expected: PASS (or only pre-existing errors unrelated to TagPicker).

- [ ] **Step 3: Commit**

```bash
git add apps/web/src/lib/ui/components/TagPicker.svelte
git commit -m "feat: add TagPicker component"
```

---

## Task 7: Frontend — TagFilter.svelte

**Files:**
- Create: `apps/web/src/lib/ui/components/TagFilter.svelte`

- [ ] **Step 1: Create `apps/web/src/lib/ui/components/TagFilter.svelte`**

```svelte
<script lang="ts">
	import TagPicker from './TagPicker.svelte';

	const {
		selected = $bindable([]),
		mode = $bindable<'AND' | 'OR'>('OR'),
		onchange
	} = $props<{
		selected?: string[];
		mode?: 'AND' | 'OR';
		onchange?: (slugs: string[], mode: 'AND' | 'OR') => void;
	}>();

	function handleTagChange(slugs: string[]) {
		selected = slugs;
		onchange?.(slugs, mode);
	}

	function setMode(m: 'AND' | 'OR') {
		mode = m;
		onchange?.(selected, m);
	}
</script>

<div class="flex flex-col gap-3">
	<div class="flex items-center gap-2">
		<span class="text-xs font-medium text-gray-500">Match:</span>
		<div class="flex rounded-md border border-gray-200 overflow-hidden text-xs">
			<button
				type="button"
				onclick={() => setMode('OR')}
				class="px-3 py-1 transition-colors {mode === 'OR'
					? 'bg-blue-600 text-white'
					: 'bg-white text-gray-600 hover:bg-gray-50'}"
			>
				Any
			</button>
			<button
				type="button"
				onclick={() => setMode('AND')}
				class="px-3 py-1 border-l border-gray-200 transition-colors {mode === 'AND'
					? 'bg-blue-600 text-white'
					: 'bg-white text-gray-600 hover:bg-gray-50'}"
			>
				All
			</button>
		</div>
	</div>

	<TagPicker bind:selected {onchange} />
</div>
```

Wait — `TagPicker` has `onchange` prop but `TagFilter` passes `onchange` which calls `handleTagChange`. Fix: `TagFilter` should use its own handler, not pass `onchange` directly through:

```svelte
<TagPicker bind:selected onchange={handleTagChange} />
```

Use this corrected version for the full file:

```svelte
<script lang="ts">
	import TagPicker from './TagPicker.svelte';

	const {
		selected = $bindable([]),
		mode = $bindable<'AND' | 'OR'>('OR'),
		onchange
	} = $props<{
		selected?: string[];
		mode?: 'AND' | 'OR';
		onchange?: (slugs: string[], mode: 'AND' | 'OR') => void;
	}>();

	function handleTagChange(slugs: string[]) {
		selected = slugs;
		onchange?.(slugs, mode);
	}

	function setMode(m: 'AND' | 'OR') {
		mode = m;
		onchange?.(selected, m);
	}
</script>

<div class="flex flex-col gap-3">
	<div class="flex items-center gap-2">
		<span class="text-xs font-medium text-gray-500">Match:</span>
		<div class="flex overflow-hidden rounded-md border border-gray-200 text-xs">
			<button
				type="button"
				onclick={() => setMode('OR')}
				class="px-3 py-1 transition-colors {mode === 'OR'
					? 'bg-blue-600 text-white'
					: 'bg-white text-gray-600 hover:bg-gray-50'}"
			>
				Any
			</button>
			<button
				type="button"
				onclick={() => setMode('AND')}
				class="px-3 py-1 border-l border-gray-200 transition-colors {mode === 'AND'
					? 'bg-blue-600 text-white'
					: 'bg-white text-gray-600 hover:bg-gray-50'}"
			>
				All
			</button>
		</div>
	</div>

	<TagPicker bind:selected onchange={handleTagChange} />
</div>
```

- [ ] **Step 2: Run svelte-check**

```bash
cd apps/web && bun run check
```

Expected: PASS.

- [ ] **Step 3: Commit**

```bash
git add apps/web/src/lib/ui/components/TagFilter.svelte
git commit -m "feat: add TagFilter component with AND/OR mode toggle"
```

---

## Task 8: Frontend — update RatingForm + client.ts

**Files:**
- Modify: `apps/web/src/lib/client/client.ts`
- Modify: `apps/web/src/lib/ui/components/RatingForm.svelte`

- [ ] **Step 1: Update `apps/web/src/lib/client/client.ts`**

Add the `TagsService` import and client:

```ts
import { createConnectTransport } from '@connectrpc/connect-web';
import { createClient } from '@connectrpc/connect';
import { RestaurantsService } from '$lib/client/generated/restaurants/v1/restaurants_service_pb';
import { UsersService } from '$lib/client/generated/users/v1/users_service_pb';
import { GoogleMapsService } from '$lib/client/generated/google_maps/v1/google_maps_service_pb';
import { AuthService } from '$lib/client/generated/auth/v1/auth_service_pb';
import { ReviewsService } from '$lib/client/generated/reviews/v1/reviews_service_pb';
import { TagsService } from '$lib/client/generated/tags/v1/tags_service_pb';

const baseUrl = import.meta.env.VITE_API_URL || 'http://localhost:3001';
const transport = createConnectTransport({
  baseUrl: baseUrl,
  useHttpGet: false,
  fetch: (input, init) => globalThis.fetch(input, { ...init, credentials: 'include' }),
  interceptors: []
});

const restaurants = createClient(RestaurantsService, transport);
const users = createClient(UsersService, transport);
const googleMaps = createClient(GoogleMapsService, transport);
const auth = createClient(AuthService, transport);
const reviews = createClient(ReviewsService, transport);
const tags = createClient(TagsService, transport);

export default { restaurants, users, googleMaps, auth, reviews, tags };
```

- [ ] **Step 2: Update `apps/web/src/lib/ui/components/RatingForm.svelte`**

Replace the entire file. Key changes: remove `tagInput`, `addTag`, `removeTag`, `handleTagKeydown`; import `TagPicker`; pass `city` and `country` to `createReview`. The component needs `city` and `country` as props.

```svelte
<script lang="ts">
	import { Button } from '$lib/components/ui/button/index.js';
	import { Label } from '$lib/components/ui/label/index.js';
	import { Star } from '@lucide/svelte';
	import client from '$lib/client/client';
	import TagPicker from './TagPicker.svelte';
	import type { ReviewProto } from '$lib/client/generated/reviews/v1/review_pb';

	const {
		googlePlacesId,
		restaurantName,
		restaurantAddress,
		city = '',
		country = '',
		existingReview,
		onSubmit
	} = $props<{
		googlePlacesId: string;
		restaurantName: string;
		restaurantAddress: string;
		city?: string;
		country?: string;
		existingReview?: ReviewProto;
		onSubmit: (review: ReviewProto) => void;
	}>();

	let rating = $state(existingReview?.rating ?? 0);
	let hoverRating = $state(0);
	let comment = $state(existingReview?.comment ?? '');
	let tags = $state<string[]>(existingReview?.tags ? [...existingReview.tags] : []);
	let loading = $state(false);
	let error = $state<string | null>(null);

	const isEdit = $derived(!!existingReview?.id);
	const displayRating = $derived(hoverRating || rating);

	async function handleSubmit() {
		if (rating < 1) {
			error = 'Please select a star rating';
			return;
		}
		error = null;
		loading = true;
		try {
			if (isEdit && existingReview) {
				const res = await client.reviews.updateReview({
					id: existingReview.id,
					comment,
					rating,
					tags
				});
				if (res.review) onSubmit(res.review);
			} else {
				const res = await client.reviews.createReview({
					googlePlacesId,
					restaurantName,
					restaurantAddress,
					city,
					country,
					comment,
					rating,
					tags
				});
				if (res.review) onSubmit(res.review);
			}
		} catch (e: unknown) {
			error = e instanceof Error ? e.message : 'Failed to save review';
		} finally {
			loading = false;
		}
	}
</script>

<div class="rounded-2xl bg-white p-6 shadow-xl">
	<h4 class="mb-4 text-base font-semibold text-gray-800">
		{isEdit ? 'Edit your rating' : 'Rate this place'}
	</h4>

	<!-- Star picker -->
	<div class="mb-4">
		<Label class="mb-1 block text-sm">Rating *</Label>
		<div class="flex gap-1">
			{#each Array(5) as _, i}
				<button
					type="button"
					onclick={() => (rating = i + 1)}
					onmouseenter={() => (hoverRating = i + 1)}
					onmouseleave={() => (hoverRating = 0)}
					class="transition-transform hover:scale-110"
					aria-label="Rate {i + 1} stars"
				>
					<Star
						class="h-7 w-7 {i < displayRating
							? 'fill-amber-400 text-amber-400'
							: 'fill-none text-gray-300'}"
					/>
				</button>
			{/each}
		</div>
	</div>

	<!-- Comment -->
	<div class="mb-4">
		<Label for="comment" class="mb-1 block text-sm">Comment (optional)</Label>
		<textarea
			id="comment"
			bind:value={comment}
			rows="3"
			placeholder="What did you think?"
			class="w-full resize-none rounded-lg border border-gray-300 px-3 py-2 text-sm text-gray-700 focus:outline-none focus:ring-2 focus:ring-blue-500"
		></textarea>
	</div>

	<!-- Tags -->
	<div class="mb-5">
		<Label class="mb-1 block text-sm">Tags (optional)</Label>
		<TagPicker bind:selected={tags} />
	</div>

	{#if error}
		<p class="mb-3 text-sm text-red-600">{error}</p>
	{/if}

	<Button onclick={handleSubmit} disabled={loading || rating < 1} class="w-full">
		{loading ? 'Saving…' : isEdit ? 'Update rating' : 'Save rating'}
	</Button>
</div>
```

- [ ] **Step 3: Run svelte-check**

```bash
cd apps/web && bun run check
```

Expected: PASS. If there are type errors about `city`/`country` in `createReview`, verify the proto regen in Task 1 succeeded and `apps/web/src/lib/client/generated/reviews/v1/reviews_service_pb.ts` contains `city` and `country` fields.

- [ ] **Step 4: Run all Go tests to confirm nothing broke**

```bash
cd apps/api && go test ./src/test/... -v
```

Expected: all PASS.

- [ ] **Step 5: Commit**

```bash
git add apps/web/src/lib/client/client.ts apps/web/src/lib/ui/components/RatingForm.svelte
git commit -m "feat: wire tags client, replace free-form tag input with TagPicker"
```

---

## Final Verification

- [ ] **Run all Go tests**

```bash
cd apps/api && go test ./src/test/... -v
```

Expected: all PASS.

- [ ] **Run Go build**

```bash
cd apps/api && go build ./...
```

Expected: PASS.

- [ ] **Run svelte-check**

```bash
cd apps/web && bun run check
```

Expected: PASS (0 errors).

---

## Notes for the Implementer

- **Proto field renumbering:** In Task 1, `restaurant.proto`'s `created_at` moves from field 6→7 and `updated_at` from 7→8 because city=5 and country=6 occupy those slots. No backwards compat needed.
- **city/country on FirstOrCreate:** `Attrs()` only sets fields on create, not update. Existing restaurants without city/country will have empty strings until they're reviewed again. This is acceptable per spec.
- **TagPicker in TagFilter:** `TagFilter` uses `bind:selected` on `TagPicker` to keep the parent's state in sync, but also passes `onchange={handleTagChange}` so `TagFilter` can intercept and re-emit with the mode.
- **Callers of RatingForm:** The `city` and `country` props are optional (default `''`). Existing callers don't need to be updated immediately — they'll just send empty strings. When `RatingForm` is used in context where a `Place` object is available (e.g. after Google Places autocomplete), pass `city` and `country` from the Place's `addressComponents`.
