# Filter & Sort for Reviews and Wishlist — Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add filter and sort parameters to `ListReviews` and `ListWishlist` — proto changes, Go backend query building, and collapsible filter panels in the SvelteKit frontend.

**Architecture:** Proto-first: add filter enums and fields to the reviews and wishlist service protos, regenerate, implement conditional GORM query building in the service handlers, then build collapsible filter panels in SvelteKit. Tags are already predefined (via `TagsService.ListTags`) and `TagPicker.svelte` already exists and is wired into `RatingForm.svelte` — reuse it unchanged for the reviews filter panel.

**Tech Stack:** Protocol Buffers + Connect-RPC, Go + GORM (PostgreSQL JSON operators, ILIKE), SvelteKit 5 runes (`$state`, `$derived`, `$effect`), ShadCN Svelte, Nx monorepo (`bunx nx`).

---

## File Map

| File | Change |
|------|--------|
| `packages/protos/reviews/v1/reviews_service.proto` | Add `ReviewSortBy` enum, `TagFilterMode` enum, filter fields to `ListReviewsRequest` |
| `packages/protos/wishlist/v1/wishlist_service.proto` | Add `WishlistSortBy` enum, filter fields to `ListWishlistRequest` |
| `apps/api/src/services/reviews_service.go` | Apply filters in `ListReviews` |
| `apps/api/src/services/wishlist_service.go` | Apply filters in `ListWishlist` |
| `apps/api/src/test/reviews_service_test.go` | Tests for invalid filter combinations |
| `apps/api/src/test/wishlist_service_test.go` | Tests for invalid filter combinations |
| `apps/web/src/routes/reviews/+page.svelte` | Add filter state + collapsible filter panel |
| `apps/web/src/routes/wishlist/+page.svelte` | Add filter state + collapsible filter panel |

**Not changing:** `TagPicker.svelte` (already complete), `RatingForm.svelte` (tags already use `TagPicker`), `client.ts` (no new services).

---

## Task 1: Proto changes + codegen

**Files:**
- Modify: `packages/protos/reviews/v1/reviews_service.proto`
- Modify: `packages/protos/wishlist/v1/wishlist_service.proto`

- [ ] **Step 1: Edit `reviews_service.proto`** — replace the entire file with:

```protobuf
syntax = "proto3";

package reviews.v1;

import "restaurants/v1/restaurant.proto";
import "reviews/v1/review.proto";

option go_package = "api/src/generated/reviews/v1";

enum ReviewSortBy {
  REVIEW_SORT_BY_UNSPECIFIED = 0;
  REVIEW_SORT_BY_DATE_DESC = 1;
  REVIEW_SORT_BY_DATE_ASC = 2;
  REVIEW_SORT_BY_RATING_DESC = 3;
  REVIEW_SORT_BY_RATING_ASC = 4;
}

enum TagFilterMode {
  TAG_FILTER_MODE_UNSPECIFIED = 0;
  TAG_FILTER_MODE_OR = 1;
  TAG_FILTER_MODE_AND = 2;
}

service ReviewsService {
  rpc CreateReview(CreateReviewRequest) returns (CreateReviewResponse);
  rpc GetReview(GetReviewRequest) returns (GetReviewResponse);
  rpc UpdateReview(UpdateReviewRequest) returns (UpdateReviewResponse);
  rpc DeleteReview(DeleteReviewRequest) returns (DeleteReviewResponse);
  rpc ListReviews(ListReviewsRequest) returns (ListReviewsResponse);
  rpc ListRestaurantReviews(ListRestaurantReviewsRequest) returns (ListRestaurantReviewsResponse);
}

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

message CreateReviewResponse {
  ReviewProto review = 1;
  restaurants.v1.RestaurantProto restaurant = 2;
}

message GetReviewRequest {
  string id = 1;
}

message GetReviewResponse {
  ReviewProto review = 1;
}

message UpdateReviewRequest {
  string id = 1;
  string comment = 2;
  double rating = 3;
  repeated string tags = 4;
}

message UpdateReviewResponse {
  ReviewProto review = 1;
}

message DeleteReviewRequest {
  string id = 1;
}

message DeleteReviewResponse {
  bool success = 1;
}

message ListReviewsRequest {
  string google_places_id = 1;
  repeated string tag_slugs = 2;
  TagFilterMode tag_filter_mode = 3;
  double min_rating = 4;
  double max_rating = 5;
  string comment_search = 6;
  string city = 7;
  string country = 8;
  ReviewSortBy sort_by = 9;
}

message ListReviewsResponse {
  repeated ReviewProto reviews = 1;
}

message ListRestaurantReviewsRequest {
  string google_places_id = 1;
}

message ListRestaurantReviewsResponse {
  repeated ReviewProto reviews = 1;
  double average_rating = 2;
  string restaurant_name = 3;
  string restaurant_address = 4;
  string restaurant_city = 5;
  string restaurant_country = 6;
}
```

- [ ] **Step 2: Edit `wishlist_service.proto`** — replace the entire file with:

```protobuf
syntax = "proto3";

package wishlist.v1;

import "wishlist/v1/wishlist_item.proto";

option go_package = "api/src/generated/wishlist/v1";

enum WishlistSortBy {
  WISHLIST_SORT_BY_UNSPECIFIED = 0;
  WISHLIST_SORT_BY_DATE_DESC = 1;
  WISHLIST_SORT_BY_DATE_ASC = 2;
  WISHLIST_SORT_BY_NAME_ASC = 3;
  WISHLIST_SORT_BY_NAME_DESC = 4;
}

service WishlistService {
  rpc AddToWishlist(AddToWishlistRequest) returns (AddToWishlistResponse);
  rpc RemoveFromWishlist(RemoveFromWishlistRequest) returns (RemoveFromWishlistResponse);
  rpc ListWishlist(ListWishlistRequest) returns (ListWishlistResponse);
}

message AddToWishlistRequest {
  string google_places_id = 1;
  string restaurant_name = 2;
  string restaurant_address = 3;
  string city = 4;
  string country = 5;
}

message AddToWishlistResponse {
  WishlistItemProto item = 1;
}

message RemoveFromWishlistRequest {
  string google_places_id = 1;
}

message RemoveFromWishlistResponse {
  bool success = 1;
}

message ListWishlistRequest {
  string google_places_id = 1;
  string city = 2;
  string country = 3;
  WishlistSortBy sort_by = 4;
}

message ListWishlistResponse {
  repeated WishlistItemProto items = 1;
}
```

- [ ] **Step 3: Run codegen**

```bash
bunx nx run protos:generate
```

Expected: No errors. Files updated in `apps/api/src/generated/reviews/v1/` and `apps/web/src/lib/client/generated/reviews/v1/` (and wishlist equivalents).

- [ ] **Step 4: Verify Go build compiles**

```bash
bunx nx run api:build
```

Expected: `Successfully ran target build for project api`

- [ ] **Step 5: Commit**

```bash
git add packages/protos/reviews/v1/reviews_service.proto packages/protos/wishlist/v1/wishlist_service.proto apps/api/src/generated apps/web/src/lib/client/generated
git commit -m "feat: add filter/sort enums and fields to ListReviews + ListWishlist protos"
```

---

## Task 2: ListReviews backend filter implementation

**Files:**
- Modify: `apps/api/src/services/reviews_service.go`
- Modify: `apps/api/src/test/reviews_service_test.go`

The `reviews` table has `tags` stored as a JSON array (e.g. `["cozy","italian"]`). Tag filtering uses `LIKE '%"slug"%'` which is reliable for well-formed slugs with no special characters. City/country live on the `restaurants` table — a JOIN is added when those filters are active. The `REVIEW_SORT_BY_UNSPECIFIED` (default `0`) value maps to `DATE_DESC`.

- [ ] **Step 1: Write the failing tests** — add to `apps/api/src/test/reviews_service_test.go`:

```go
func TestReviewsService_ListReviews_InvalidRatingRange(t *testing.T) {
	svc := &services.ReviewsService{}
	req := connect.NewRequest(&reviewsv1.ListReviewsRequest{
		MinRating: 4,
		MaxRating: 2,
	})
	_, err := svc.ListReviews(context.Background(), req)
	if err == nil {
		t.Fatal("expected error for min_rating > max_rating, got nil")
	}
	var connectErr *connect.Error
	if !errors.As(err, &connectErr) || connectErr.Code() != connect.CodeInvalidArgument {
		t.Fatalf("expected CodeInvalidArgument, got %v", err)
	}
}

func TestReviewsService_ListReviews_NilDB(t *testing.T) {
	svc := &services.ReviewsService{}
	req := connect.NewRequest(&reviewsv1.ListReviewsRequest{})
	_, err := svc.ListReviews(context.Background(), req)
	if err == nil {
		t.Fatal("expected error from nil DB, got nil")
	}
}
```

Note: you'll also need `"errors"` in the import block if it's not already there. Check the existing imports — `"connectrpc.com/connect"` and `"context"` are already imported.

- [ ] **Step 2: Run tests to confirm they fail**

```bash
cd apps/api && go test ./src/test/... -run "TestReviewsService_ListReviews" -v
```

Expected: `FAIL` — `ListReviews` doesn't validate rating range yet, so `TestReviewsService_ListReviews_InvalidRatingRange` passes vacuously (no error returned) → test fails. `TestReviewsService_ListReviews_NilDB` may pass depending on current code.

- [ ] **Step 3: Replace `ListReviews` in `apps/api/src/services/reviews_service.go`**

Add `"fmt"` and `"strings"` to the import block (they aren't imported yet). The full import block should be:

```go
import (
	restaurantspb "api/src/generated/restaurants/v1"
	v1 "api/src/generated/reviews/v1"
	"api/src/generated/reviews/v1/v1connect"
	"api/src/internal/models"
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"connectrpc.com/connect"
	"github.com/valkey-io/valkey-go"
	"gorm.io/gorm"
)
```

Replace the entire `ListReviews` method with:

```go
func (s *ReviewsService) ListReviews(
	ctx context.Context,
	req *connect.Request[v1.ListReviewsRequest],
) (*connect.Response[v1.ListReviewsResponse], error) {
	if s.DB == nil {
		return nil, connect.NewError(connect.CodeInternal, errors.New("database not initialized"))
	}

	userID, err := getUserIDFromSession(ctx, req.Header(), s.Valkey)
	if err != nil {
		return nil, err
	}

	// Validate rating range when both are set
	if req.Msg.MinRating > 0 && req.Msg.MaxRating > 0 && req.Msg.MinRating > req.Msg.MaxRating {
		return nil, connect.NewError(connect.CodeInvalidArgument, errors.New("min_rating must not exceed max_rating"))
	}

	needsRestaurantJoin := req.Msg.City != "" || req.Msg.Country != ""

	query := s.DB.WithContext(ctx).Preload("User").Where("reviews.user_id = ?", userID)

	if needsRestaurantJoin {
		query = query.Joins("JOIN restaurants ON restaurants.id = reviews.restaurant_id").
			Preload("Restaurant")
	} else {
		query = query.Preload("Restaurant")
	}

	if req.Msg.GooglePlacesId != "" {
		query = query.Where("reviews.google_places_id = ?", req.Msg.GooglePlacesId)
	}

	// Tag filter
	if len(req.Msg.TagSlugs) > 0 {
		if req.Msg.TagFilterMode == v1.TagFilterMode_TAG_FILTER_MODE_AND {
			// AND: every specified tag must appear in the JSON array
			for _, slug := range req.Msg.TagSlugs {
				query = query.Where("reviews.tags LIKE ?", fmt.Sprintf(`%%"%s"%%`, slug))
			}
		} else {
			// OR (default): at least one specified tag must appear
			conditions := make([]string, len(req.Msg.TagSlugs))
			args := make([]interface{}, len(req.Msg.TagSlugs))
			for i, slug := range req.Msg.TagSlugs {
				conditions[i] = "reviews.tags LIKE ?"
				args[i] = fmt.Sprintf(`%%"%s"%%`, slug)
			}
			query = query.Where(strings.Join(conditions, " OR "), args...)
		}
	}

	// Rating range
	if req.Msg.MinRating > 0 {
		query = query.Where("reviews.rating >= ?", req.Msg.MinRating)
	}
	if req.Msg.MaxRating > 0 {
		query = query.Where("reviews.rating <= ?", req.Msg.MaxRating)
	}

	// Comment keyword search (case-insensitive)
	if req.Msg.CommentSearch != "" {
		query = query.Where("reviews.comment ILIKE ?", "%"+req.Msg.CommentSearch+"%")
	}

	// City / country filter (requires restaurant join)
	if req.Msg.City != "" {
		query = query.Where("restaurants.city ILIKE ?", "%"+req.Msg.City+"%")
	}
	if req.Msg.Country != "" {
		query = query.Where("restaurants.country ILIKE ?", "%"+req.Msg.Country+"%")
	}

	// Sort order
	switch req.Msg.SortBy {
	case v1.ReviewSortBy_REVIEW_SORT_BY_DATE_ASC:
		query = query.Order("reviews.created_at ASC")
	case v1.ReviewSortBy_REVIEW_SORT_BY_RATING_DESC:
		query = query.Order("reviews.rating DESC")
	case v1.ReviewSortBy_REVIEW_SORT_BY_RATING_ASC:
		query = query.Order("reviews.rating ASC")
	default: // UNSPECIFIED and DATE_DESC both → newest first
		query = query.Order("reviews.created_at DESC")
	}

	var reviews []models.Review
	if err := query.Find(&reviews).Error; err != nil {
		return nil, err
	}

	protos := make([]*v1.ReviewProto, len(reviews))
	for i, r := range reviews {
		protos[i] = r.ToProto()
	}

	return connect.NewResponse(&v1.ListReviewsResponse{Reviews: protos}), nil
}
```

- [ ] **Step 4: Run tests — confirm they pass**

```bash
cd apps/api && go test ./src/test/... -run "TestReviewsService_ListReviews" -v
```

Expected: Both tests `PASS`.

- [ ] **Step 5: Run full Go build to confirm no regressions**

```bash
cd /path/to/repo && bunx nx run api:build
```

Expected: `Successfully ran target build for project api`

- [ ] **Step 6: Commit**

```bash
git add apps/api/src/services/reviews_service.go apps/api/src/test/reviews_service_test.go
git commit -m "feat: apply filter + sort params in ListReviews"
```

---

## Task 3: ListWishlist backend filter implementation

**Files:**
- Modify: `apps/api/src/services/wishlist_service.go`
- Modify: `apps/api/src/test/wishlist_service_test.go`

City/country and name-sort require joining the `restaurants` table. The JOIN is added conditionally.

- [ ] **Step 1: Write the failing test** — add to `apps/api/src/test/wishlist_service_test.go`:

```go
func TestWishlistService_ListWishlist_NilDB_WithFilters(t *testing.T) {
	svc := &services.WishlistService{}
	req := connect.NewRequest(&wishlistv1.ListWishlistRequest{
		City:   "Paris",
		SortBy: wishlistv1.WishlistSortBy_WISHLIST_SORT_BY_NAME_ASC,
	})
	_, err := svc.ListWishlist(context.Background(), req)
	if err == nil {
		t.Fatal("expected error from nil DB, got nil")
	}
}
```

- [ ] **Step 2: Run test to confirm it fails**

```bash
cd apps/api && go test ./src/test/... -run "TestWishlistService_ListWishlist_NilDB_WithFilters" -v
```

Expected: `FAIL` — `WishlistSortBy_WISHLIST_SORT_BY_NAME_ASC` doesn't exist yet (or the service doesn't handle the filter).

- [ ] **Step 3: Replace `ListWishlist` in `apps/api/src/services/wishlist_service.go`**

Add `"strings"` to the import block if not present. The new import block:

```go
import (
	wishlistv1 "api/src/generated/wishlist/v1"
	"api/src/generated/wishlist/v1/v1connect"
	"api/src/internal/models"
	"context"
	"errors"
	"strings"

	"connectrpc.com/connect"
	"github.com/valkey-io/valkey-go"
	"gorm.io/gorm"
)
```

Replace the entire `ListWishlist` method with:

```go
func (s *WishlistService) ListWishlist(
	ctx context.Context,
	req *connect.Request[wishlistv1.ListWishlistRequest],
) (*connect.Response[wishlistv1.ListWishlistResponse], error) {
	if s.DB == nil {
		return nil, connect.NewError(connect.CodeInternal, errors.New("database not initialized"))
	}

	userID, err := getUserIDFromSession(ctx, req.Header(), s.Valkey)
	if err != nil {
		return nil, err
	}

	needsJoin := req.Msg.City != "" || req.Msg.Country != "" ||
		req.Msg.SortBy == wishlistv1.WishlistSortBy_WISHLIST_SORT_BY_NAME_ASC ||
		req.Msg.SortBy == wishlistv1.WishlistSortBy_WISHLIST_SORT_BY_NAME_DESC

	query := s.DB.WithContext(ctx).
		Preload("Restaurant").
		Where("wishlist_items.user_id = ?", userID)

	if needsJoin {
		query = query.Joins("JOIN restaurants ON restaurants.id = wishlist_items.restaurant_id")
	}

	if req.Msg.GooglePlacesId != "" {
		query = query.Where("wishlist_items.google_places_id = ?", req.Msg.GooglePlacesId)
	}

	if req.Msg.City != "" {
		query = query.Where("restaurants.city ILIKE ?", "%"+req.Msg.City+"%")
	}
	if req.Msg.Country != "" {
		query = query.Where("restaurants.country ILIKE ?", "%"+req.Msg.Country+"%")
	}

	switch req.Msg.SortBy {
	case wishlistv1.WishlistSortBy_WISHLIST_SORT_BY_DATE_ASC:
		query = query.Order("wishlist_items.created_at ASC")
	case wishlistv1.WishlistSortBy_WISHLIST_SORT_BY_NAME_ASC:
		query = query.Order("restaurants.name ASC")
	case wishlistv1.WishlistSortBy_WISHLIST_SORT_BY_NAME_DESC:
		query = query.Order("restaurants.name DESC")
	default: // UNSPECIFIED and DATE_DESC → newest first
		query = query.Order("wishlist_items.created_at DESC")
	}

	var items []models.WishlistItem
	if err := query.Find(&items).Error; err != nil {
		return nil, err
	}

	protos := make([]*wishlistv1.WishlistItemProto, len(items))
	for i, item := range items {
		protos[i] = item.ToProto()
	}

	return connect.NewResponse(&wishlistv1.ListWishlistResponse{Items: protos}), nil
}
```

Also remove the now-unused `"strings"` import if it ended up unused (check — it's not used in `ListWishlist` so don't add it). The import block just needs:

```go
import (
	wishlistv1 "api/src/generated/wishlist/v1"
	"api/src/generated/wishlist/v1/v1connect"
	"api/src/internal/models"
	"context"
	"errors"

	"connectrpc.com/connect"
	"github.com/valkey-io/valkey-go"
	"gorm.io/gorm"
)
```

- [ ] **Step 4: Run tests**

```bash
cd apps/api && go test ./src/test/... -run "TestWishlistService" -v
```

Expected: All `TestWishlistService_*` tests pass.

- [ ] **Step 5: Run full Go build**

```bash
bunx nx run api:build
```

Expected: `Successfully ran target build for project api`

- [ ] **Step 6: Run all API tests**

```bash
bunx nx run api:test
```

Expected: `Successfully ran target test for project api`

- [ ] **Step 7: Commit**

```bash
git add apps/api/src/services/wishlist_service.go apps/api/src/test/wishlist_service_test.go
git commit -m "feat: apply filter + sort params in ListWishlist"
```

---

## Task 4: Reviews page filter panel

**Files:**
- Modify: `apps/web/src/routes/reviews/+page.svelte`

Design: a "Filters" toggle button at the top. When open, shows a filter panel with tags (reuses `TagPicker`), AND/OR toggle, min/max rating selects, comment keyword search (debounced 300 ms), city/country inputs, and sort dropdown. Filter state is reactive — any change re-fetches. A badge on the button shows how many filters are active. A "Clear" button resets everything.

Generated TypeScript enums: `ReviewSortBy` and `TagFilterMode` come from `$lib/client/generated/reviews/v1/reviews_service_pb`.

- [ ] **Step 1: Replace `apps/web/src/routes/reviews/+page.svelte`** with:

```svelte
<script lang="ts">
	import { onMount } from 'svelte';
	import { goto } from '$app/navigation';
	import { auth } from '$lib/state/auth.svelte';
	import client from '$lib/client/client';
	import { ReviewSortBy, TagFilterMode } from '$lib/client/generated/reviews/v1/reviews_service_pb';
	import type { ReviewProto } from '$lib/client/generated/reviews/v1/review_pb';
	import { Star } from '@lucide/svelte';
	import { Button } from '$lib/components/ui/button/index.js';
	import RatingForm from '$lib/ui/components/RatingForm.svelte';
	import TagPicker from '$lib/ui/components/TagPicker.svelte';

	let reviews = $state<ReviewProto[]>([]);
	let loading = $state(true);
	let editingId = $state<string | null>(null);
	let deleting = $state<Set<string>>(new Set());
	let mounted = $state(false);

	// Filter state
	let showFilters = $state(false);
	let tagSlugs = $state<string[]>([]);
	let tagMode = $state<'or' | 'and'>('or');
	let minRating = $state(0);
	let maxRating = $state(0);
	let commentRaw = $state('');
	let commentSearch = $state('');
	let city = $state('');
	let country = $state('');
	let sortBy = $state('date-desc');

	// Debounce comment search
	$effect(() => {
		const val = commentRaw;
		const id = setTimeout(() => {
			commentSearch = val;
		}, 300);
		return () => clearTimeout(id);
	});

	let activeFilterCount = $derived(
		(tagSlugs.length > 0 ? 1 : 0) +
			(minRating > 0 || maxRating > 0 ? 1 : 0) +
			(commentSearch.trim() !== '' ? 1 : 0) +
			(city.trim() !== '' ? 1 : 0) +
			(country.trim() !== '' ? 1 : 0) +
			(sortBy !== 'date-desc' ? 1 : 0)
	);

	function clearFilters() {
		tagSlugs = [];
		tagMode = 'or';
		minRating = 0;
		maxRating = 0;
		commentRaw = '';
		commentSearch = '';
		city = '';
		country = '';
		sortBy = 'date-desc';
	}

	function toSortByEnum(s: string): ReviewSortBy {
		switch (s) {
			case 'date-asc':
				return ReviewSortBy.REVIEW_SORT_BY_DATE_ASC;
			case 'rating-desc':
				return ReviewSortBy.REVIEW_SORT_BY_RATING_DESC;
			case 'rating-asc':
				return ReviewSortBy.REVIEW_SORT_BY_RATING_ASC;
			default:
				return ReviewSortBy.REVIEW_SORT_BY_DATE_DESC;
		}
	}

	async function loadReviews() {
		loading = true;
		try {
			const res = await client.reviews.listReviews({
				tagSlugs,
				tagFilterMode:
					tagMode === 'and' ? TagFilterMode.TAG_FILTER_MODE_AND : TagFilterMode.TAG_FILTER_MODE_OR,
				minRating,
				maxRating,
				commentSearch,
				city,
				country,
				sortBy: toSortByEnum(sortBy)
			});
			reviews = res.reviews ?? [];
		} catch (e) {
			console.error('Failed to load reviews:', e);
		} finally {
			loading = false;
		}
	}

	// Reactive reload when any filter changes (only after mount + auth confirmed)
	$effect(() => {
		if (!mounted) return;
		// Read all filter deps so $effect subscribes to them
		void [tagSlugs, tagMode, minRating, maxRating, commentSearch, city, country, sortBy];
		loadReviews();
	});

	async function deleteReview(id: string) {
		deleting = new Set([...deleting, id]);
		const removed = reviews.find((r) => r.id === id)!;
		reviews = reviews.filter((r) => r.id !== id);
		try {
			await client.reviews.deleteReview({ id });
		} catch (e) {
			console.error('Failed to delete review:', e);
			reviews = [...reviews, removed];
		} finally {
			deleting.delete(id);
			deleting = new Set(deleting);
		}
	}

	onMount(() => {
		if (!auth.isLoggedIn) {
			goto('/?login=1');
			return;
		}
		mounted = true;
		// initial load is triggered by $effect above when mounted becomes true
	});
</script>

<div class="container mx-auto max-w-3xl space-y-6 p-6">
	<h2 class="text-2xl font-semibold text-blue-800">My Reviews</h2>

	<!-- Filter bar -->
	<div class="space-y-3">
		<div class="flex flex-wrap items-center gap-2">
			<Button
				variant={showFilters ? 'default' : 'outline'}
				size="sm"
				onclick={() => (showFilters = !showFilters)}
			>
				Filters{activeFilterCount > 0 ? ` (${activeFilterCount})` : ''}
			</Button>
			{#if activeFilterCount > 0}
				<Button variant="ghost" size="sm" onclick={clearFilters}>Clear all</Button>
			{/if}
			<div class="ml-auto flex items-center gap-2">
				<label for="sort" class="text-sm text-gray-600">Sort:</label>
				<select
					id="sort"
					bind:value={sortBy}
					class="rounded-md border border-gray-300 px-2 py-1 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500"
				>
					<option value="date-desc">Newest first</option>
					<option value="date-asc">Oldest first</option>
					<option value="rating-desc">Highest rated</option>
					<option value="rating-asc">Lowest rated</option>
				</select>
			</div>
		</div>

		{#if showFilters}
			<div class="space-y-4 rounded-lg border border-gray-200 bg-gray-50 p-4">
				<!-- Tags -->
				<div>
					<div class="mb-1 flex items-center gap-3">
						<span class="text-sm font-medium text-gray-700">Tags</span>
						<div class="flex items-center gap-1 rounded-full bg-white border border-gray-300 p-0.5">
							<button
								type="button"
								onclick={() => (tagMode = 'or')}
								class="rounded-full px-2.5 py-0.5 text-xs font-medium transition-colors {tagMode === 'or' ? 'bg-blue-600 text-white' : 'text-gray-600 hover:bg-gray-100'}"
							>
								Any (OR)
							</button>
							<button
								type="button"
								onclick={() => (tagMode = 'and')}
								class="rounded-full px-2.5 py-0.5 text-xs font-medium transition-colors {tagMode === 'and' ? 'bg-blue-600 text-white' : 'text-gray-600 hover:bg-gray-100'}"
							>
								All (AND)
							</button>
						</div>
					</div>
					<TagPicker bind:selected={tagSlugs} />
				</div>

				<!-- Rating range -->
				<div class="flex flex-wrap items-center gap-2">
					<span class="text-sm font-medium text-gray-700">Rating:</span>
					<select
						bind:value={minRating}
						class="rounded-md border border-gray-300 px-2 py-1 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500"
					>
						<option value={0}>Min ★</option>
						{#each [1, 2, 3, 4, 5] as n}
							<option value={n}>{n} ★</option>
						{/each}
					</select>
					<span class="text-sm text-gray-500">to</span>
					<select
						bind:value={maxRating}
						class="rounded-md border border-gray-300 px-2 py-1 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500"
					>
						<option value={0}>Max ★</option>
						{#each [1, 2, 3, 4, 5] as n}
							<option value={n}>{n} ★</option>
						{/each}
					</select>
				</div>

				<!-- Comment search -->
				<div>
					<label for="comment-search" class="mb-1 block text-sm font-medium text-gray-700">
						Comment contains
					</label>
					<input
						id="comment-search"
						type="text"
						bind:value={commentRaw}
						placeholder="Search in comments…"
						class="w-full rounded-md border border-gray-300 px-3 py-1.5 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500"
					/>
				</div>

				<!-- City + Country -->
				<div class="grid grid-cols-2 gap-3">
					<div>
						<label for="filter-city" class="mb-1 block text-sm font-medium text-gray-700">City</label>
						<input
							id="filter-city"
							type="text"
							bind:value={city}
							placeholder="e.g. Paris"
							class="w-full rounded-md border border-gray-300 px-3 py-1.5 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500"
						/>
					</div>
					<div>
						<label for="filter-country" class="mb-1 block text-sm font-medium text-gray-700">Country</label>
						<input
							id="filter-country"
							type="text"
							bind:value={country}
							placeholder="e.g. France"
							class="w-full rounded-md border border-gray-300 px-3 py-1.5 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500"
						/>
					</div>
				</div>
			</div>
		{/if}
	</div>

	{#if !auth.isLoggedIn}
		<p class="text-sm text-gray-500">Please sign in to view your reviews.</p>
	{:else if loading}
		<div class="flex items-center gap-2 text-sm text-gray-500">
			<div class="h-4 w-4 animate-spin rounded-full border-2 border-gray-300 border-t-blue-500"></div>
			Loading…
		</div>
	{:else if reviews.length === 0}
		<p class="text-sm text-gray-500">
			{#if activeFilterCount > 0}
				No reviews match the current filters. <button
					type="button"
					onclick={clearFilters}
					class="text-blue-600 underline hover:no-underline">Clear filters</button
				>
			{:else}
				No reviews yet. Search for a restaurant on the <a href="/" class="text-blue-600 hover:underline"
					>home page</a
				> to leave one.
			{/if}
		</p>
	{:else}
		<ul class="space-y-3">
			{#each reviews as review (review.id)}
				<li class="rounded-lg border border-gray-200 bg-white p-4 shadow-sm">
					{#if editingId === review.id}
						<RatingForm
							googlePlacesId={review.googlePlacesId}
							restaurantName={review.restaurantName}
							restaurantAddress={review.restaurantAddress}
							existingReview={review}
							onSubmit={(updated) => {
								reviews = reviews.map((r) => (r.id === updated.id ? updated : r));
								editingId = null;
							}}
						/>
						<Button variant="ghost" size="sm" class="mt-2" onclick={() => (editingId = null)}>
							Cancel
						</Button>
					{:else}
						<div class="mb-2 flex items-start justify-between gap-2">
							<div class="min-w-0">
								{#if review.restaurantName}
									{#if review.googlePlacesId}
										<a
											href="/restaurants/{encodeURIComponent(review.googlePlacesId)}"
											class="truncate font-medium text-blue-700 hover:underline"
										>{review.restaurantName}</a>
									{:else}
										<p class="truncate font-medium text-gray-900">{review.restaurantName}</p>
									{/if}
								{/if}
								{#if review.restaurantAddress}
									<p class="truncate text-sm text-gray-500">{review.restaurantAddress}</p>
								{/if}
								{#if review.restaurantCity || review.restaurantCountry}
									<p class="text-xs text-gray-400">
										{[review.restaurantCity, review.restaurantCountry].filter(Boolean).join(', ')}
									</p>
								{/if}
							</div>
							<div class="flex shrink-0 gap-1">
								<Button variant="outline" size="sm" onclick={() => (editingId = review.id)}>
									Edit
								</Button>
								<Button
									variant="outline"
									size="sm"
									disabled={deleting.has(review.id)}
									onclick={() => deleteReview(review.id)}
									class="text-red-600 hover:border-red-300 hover:text-red-700"
								>
									{deleting.has(review.id) ? 'Deleting…' : 'Delete'}
								</Button>
							</div>
						</div>

						<div class="mb-2 flex items-center gap-2">
							<div class="flex items-center gap-0.5">
								{#each Array(5) as _, i}
									<Star
										class="h-4 w-4 {i < review.rating
											? 'fill-amber-400 text-amber-400'
											: 'fill-none text-gray-300'}"
									/>
								{/each}
							</div>
							<span class="text-sm font-semibold text-gray-800">{review.rating.toFixed(1)}</span>
						</div>

						{#if review.comment}
							<p class="mb-2 text-sm leading-relaxed text-gray-600">{review.comment}</p>
						{/if}

						{#if review.tags && review.tags.length > 0}
							<div class="flex flex-wrap gap-1.5">
								{#each review.tags as tag}
									<span
										class="rounded-full bg-blue-100 px-2.5 py-0.5 text-xs font-medium text-blue-700"
									>
										{tag}
									</span>
								{/each}
							</div>
						{/if}
					{/if}
				</li>
			{/each}
		</ul>
	{/if}
</div>
```

- [ ] **Step 2: Run svelte-check**

```bash
bunx nx run web:check
```

Expected: `svelte-check found 0 errors and 0 warnings`

- [ ] **Step 3: Commit**

```bash
git add apps/web/src/routes/reviews/+page.svelte
git commit -m "feat: filter panel on reviews page (tags, rating, comment, city/country, sort)"
```

---

## Task 5: Wishlist page filter panel

**Files:**
- Modify: `apps/web/src/routes/wishlist/+page.svelte`

Simpler than reviews: only city/country inputs and a sort dropdown.

- [ ] **Step 1: Add filter imports and state** — edit `apps/web/src/routes/wishlist/+page.svelte`. Replace the `<script>` block with:

```svelte
<script lang="ts">
	import { onMount } from 'svelte';
	import { goto } from '$app/navigation';
	import { auth } from '$lib/state/auth.svelte';
	import client from '$lib/client/client';
	import { WishlistSortBy } from '$lib/client/generated/wishlist/v1/wishlist_service_pb';
	import type { WishlistItemProto } from '$lib/client/generated/wishlist/v1/wishlist_item_pb';
	import type { ReviewProto } from '$lib/client/generated/reviews/v1/review_pb';
	import type { Place } from '$lib/client/generated/google_maps/v1/google_maps_service_pb';
	import { Button } from '$lib/components/ui/button/index.js';
	import ExpandableRestaurantInfo from '$lib/ui/components/ExpandableRestaurantInfo.svelte';
	import RatingForm from '$lib/ui/components/RatingForm.svelte';
	import RestaurantSearch from '$lib/ui/components/RestaurantSearch.svelte';

	let items = $state<WishlistItemProto[]>([]);
	let loading = $state(true);
	let removing = $state<Set<string>>(new Set());
	let ratingId = $state<string | null>(null);
	let mounted = $state(false);

	let searchedPlace = $state<Place | null>(null);
	let searchAction = $state<'review' | null>(null);
	let savingToWishlist = $state(false);

	// Filter state
	let city = $state('');
	let country = $state('');
	let sortBy = $state('date-desc');

	let activeFilterCount = $derived(
		(city.trim() !== '' ? 1 : 0) +
			(country.trim() !== '' ? 1 : 0) +
			(sortBy !== 'date-desc' ? 1 : 0)
	);

	function clearFilters() {
		city = '';
		country = '';
		sortBy = 'date-desc';
	}

	function toSortByEnum(s: string): WishlistSortBy {
		switch (s) {
			case 'date-asc':
				return WishlistSortBy.WISHLIST_SORT_BY_DATE_ASC;
			case 'name-asc':
				return WishlistSortBy.WISHLIST_SORT_BY_NAME_ASC;
			case 'name-desc':
				return WishlistSortBy.WISHLIST_SORT_BY_NAME_DESC;
			default:
				return WishlistSortBy.WISHLIST_SORT_BY_DATE_DESC;
		}
	}

	function handleSearchSelect(place: Place) {
		searchedPlace = place;
		searchAction = null;
	}

	async function saveToWishlist() {
		if (!searchedPlace) return;
		savingToWishlist = true;
		try {
			await client.wishlist.addToWishlist({
				googlePlacesId: searchedPlace.name || '',
				restaurantName: searchedPlace.displayName?.text || '',
				restaurantAddress: searchedPlace.formattedAddress || '',
				city: searchedPlace.postalAddress?.locality ?? '',
				country: searchedPlace.postalAddress?.country ?? ''
			});
			await loadWishlist();
			searchedPlace = null;
		} catch (e) {
			console.error('Failed to add to wishlist:', e);
		} finally {
			savingToWishlist = false;
		}
	}

	function handleSearchReview(review: ReviewProto) {
		items = items.filter((i) => i.googlePlacesId !== review.googlePlacesId);
		searchedPlace = null;
		searchAction = null;
	}

	async function loadWishlist() {
		loading = true;
		try {
			const res = await client.wishlist.listWishlist({
				city,
				country,
				sortBy: toSortByEnum(sortBy)
			});
			items = res.items ?? [];
		} catch (e) {
			console.error('Failed to load wishlist:', e);
		} finally {
			loading = false;
		}
	}

	async function remove(googlePlacesId: string) {
		removing = new Set([...removing, googlePlacesId]);
		try {
			await client.wishlist.removeFromWishlist({ googlePlacesId });
			items = items.filter((i) => i.googlePlacesId !== googlePlacesId);
		} catch (e) {
			console.error('Failed to remove from wishlist:', e);
		} finally {
			removing.delete(googlePlacesId);
			removing = new Set(removing);
		}
	}

	// Reactive reload when filters change (only after auth + mount)
	$effect(() => {
		if (!mounted) return;
		void [city, country, sortBy];
		loadWishlist();
	});

	onMount(() => {
		if (!auth.isLoggedIn) {
			goto('/?login=1');
			return;
		}
		mounted = true;
	});
</script>
```

- [ ] **Step 2: Replace the template section of `wishlist/+page.svelte`** — keep the existing `<div class="container...">` structure but add a filter bar between `<h2>` and the search section:

```svelte
<div class="container mx-auto max-w-3xl space-y-6 p-6">
	<h2 class="text-2xl font-semibold text-blue-800">My Wishlist</h2>

	<!-- Filter bar -->
	<div class="flex flex-wrap items-center gap-2">
		{#if activeFilterCount > 0}
			<Button variant="ghost" size="sm" onclick={clearFilters}>Clear filters ({activeFilterCount})</Button>
		{/if}
		<div class="ml-auto flex items-center gap-2">
			<label for="wishlist-sort" class="text-sm text-gray-600">Sort:</label>
			<select
				id="wishlist-sort"
				bind:value={sortBy}
				class="rounded-md border border-gray-300 px-2 py-1 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500"
			>
				<option value="date-desc">Newest first</option>
				<option value="date-asc">Oldest first</option>
				<option value="name-asc">Name A–Z</option>
				<option value="name-desc">Name Z–A</option>
			</select>
		</div>
		<div class="flex w-full gap-3">
			<div class="flex-1">
				<input
					type="text"
					bind:value={city}
					placeholder="Filter by city…"
					class="w-full rounded-md border border-gray-300 px-3 py-1.5 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500"
				/>
			</div>
			<div class="flex-1">
				<input
					type="text"
					bind:value={country}
					placeholder="Filter by country…"
					class="w-full rounded-md border border-gray-300 px-3 py-1.5 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500"
				/>
			</div>
		</div>
	</div>

	<section class="space-y-3">
		<h3 class="text-lg font-medium text-gray-800">Find a restaurant</h3>
		<RestaurantSearch
			placeholder="Search to add to wishlist or review…"
			onSelect={handleSearchSelect}
		/>
		{#if searchedPlace}
			<div class="space-y-3 rounded-lg border border-gray-200 bg-white p-4 shadow-sm">
				<div>
					<p class="font-medium text-gray-900">
						{searchedPlace.displayName?.text || searchedPlace.name || ''}
					</p>
					<p class="text-sm text-gray-500">{searchedPlace.formattedAddress || ''}</p>
				</div>

				{#if !searchAction}
					<div class="flex gap-2">
						<Button onclick={saveToWishlist} disabled={savingToWishlist}>
							{savingToWishlist ? 'Saving…' : '☆ Save to wishlist'}
						</Button>
						<Button variant="secondary" onclick={() => (searchAction = 'review')}>
							📝 Add review
						</Button>
						<Button variant="ghost" onclick={() => (searchedPlace = null)}>Cancel</Button>
					</div>
				{:else if searchAction === 'review'}
					<RatingForm
						googlePlacesId={searchedPlace.name || ''}
						restaurantName={searchedPlace.displayName?.text || ''}
						restaurantAddress={searchedPlace.formattedAddress || ''}
						onSubmit={handleSearchReview}
					/>
					<Button variant="ghost" size="sm" onclick={() => (searchAction = null)}>Back</Button>
				{/if}
			</div>
		{/if}
	</section>

	{#if loading}
		<div class="flex items-center gap-2 text-sm text-gray-500">
			<div
				class="h-4 w-4 animate-spin rounded-full border-2 border-gray-300 border-t-blue-500"
			></div>
			Loading…
		</div>
	{:else if items.length === 0}
		<p class="text-sm text-gray-500">
			{#if activeFilterCount > 0}
				No wishlist items match the current filters. <button
					type="button"
					onclick={clearFilters}
					class="text-blue-600 underline hover:no-underline">Clear filters</button
				>
			{:else}
				Your wishlist is empty. Search for a restaurant above to add one.
			{/if}
		</p>
	{:else}
		<ul class="space-y-3">
			{#each items as item (item.id)}
				<li class="space-y-3 rounded-lg border border-gray-200 bg-white p-4 shadow-sm">
					<ExpandableRestaurantInfo
						googlePlacesId={item.googlePlacesId}
						name={item.restaurantName}
						address={item.restaurantAddress}
						city={item.city}
						country={item.country}
					/>

					{#if ratingId !== item.id}
						<div class="flex gap-2 border-t border-gray-100 pt-1">
							<Button
								variant="outline"
								size="sm"
								class="text-red-600 hover:border-red-300 hover:text-red-700"
								disabled={removing.has(item.googlePlacesId)}
								onclick={() => remove(item.googlePlacesId)}
							>
								{removing.has(item.googlePlacesId) ? 'Removing…' : 'Remove'}
							</Button>
							<Button variant="secondary" size="sm" onclick={() => (ratingId = item.id)}>
								Rate this place
							</Button>
						</div>
					{:else}
						<div class="space-y-3 border-t border-gray-100 pt-2">
							<RatingForm
								googlePlacesId={item.googlePlacesId}
								restaurantName={item.restaurantName}
								restaurantAddress={item.restaurantAddress}
								onSubmit={() => {
									items = items.filter((i) => i.googlePlacesId !== item.googlePlacesId);
									ratingId = null;
								}}
							/>
							<Button variant="ghost" size="sm" onclick={() => (ratingId = null)}>Cancel</Button>
						</div>
					{/if}
				</li>
			{/each}
		</ul>
	{/if}
</div>
```

- [ ] **Step 3: Run svelte-check**

```bash
bunx nx run web:check
```

Expected: `svelte-check found 0 errors and 0 warnings`

- [ ] **Step 4: Commit**

```bash
git add apps/web/src/routes/wishlist/+page.svelte
git commit -m "feat: filter panel on wishlist page (city/country, sort)"
```

---

## Task 6: Final verification + format

- [ ] **Step 1: Run formatter**

```bash
bun run format
```

Expected: Files reformatted (or unchanged if already correct).

- [ ] **Step 2: Run svelte-check**

```bash
bunx nx run web:check
```

Expected: `svelte-check found 0 errors and 0 warnings`

- [ ] **Step 3: Run Go tests**

```bash
bunx nx run api:test
```

Expected: `Successfully ran target test for project api`

- [ ] **Step 4: Run Go build**

```bash
bunx nx run api:build
```

Expected: `Successfully ran target build for project api`

- [ ] **Step 5: Commit any formatting changes**

```bash
git add -A
git status
# If there are changes:
git commit -m "chore: format after filter/sort feature"
```
