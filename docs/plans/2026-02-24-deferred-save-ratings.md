# Deferred Save + Ratings Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Stop saving restaurants on search; save only when a logged-in user submits a rating (stars + comment + free-form tags) via a single atomic `CreateReview` RPC.

**Architecture:** `CreateReview` on the Go backend does `FirstOrCreate` restaurant by `google_places_id`, then creates the review in a transaction. Frontend shows a Google Places preview card + rating form after search selection. Auth gate on the search input (must be logged in). Tags stored as `[]string` with GORM JSON serializer.

**Tech Stack:** Go/GORM/ConnectRPC, SvelteKit 5 runes, shadcn-svelte Button/Label, `@lucide/svelte`, Protobuf/buf codegen.

---

## Task 1: Update review.proto

**Files:**
- Modify: `packages/protos/reviews/v1/review.proto`

**Step 1: Replace the entire file**

```protobuf
syntax = "proto3";

package reviews.v1;

option go_package = "api/src/generated/reviews/v1";

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

Change: added `repeated string tags = 9`.

---

## Task 2: Update reviews_service.proto

**Files:**
- Modify: `packages/protos/reviews/v1/reviews_service.proto`

**Step 1: Replace the entire file**

```protobuf
syntax = "proto3";

package reviews.v1;

import "reviews/v1/review.proto";
import "restaurants/v1/restaurant.proto";

option go_package = "api/src/generated/reviews/v1";

service ReviewsService {
  rpc CreateReview(CreateReviewRequest) returns (CreateReviewResponse);
  rpc GetReview(GetReviewRequest) returns (GetReviewResponse);
  rpc UpdateReview(UpdateReviewRequest) returns (UpdateReviewResponse);
  rpc DeleteReview(DeleteReviewRequest) returns (DeleteReviewResponse);
  rpc ListReviews(ListReviewsRequest) returns (ListReviewsResponse);
}

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
}

message ListReviewsResponse {
  repeated ReviewProto reviews = 1;
}
```

Changes: `CreateReviewRequest` now takes Google Places data instead of a DB restaurant ID; added `restaurants.v1.RestaurantProto` to `CreateReviewResponse`; simplified `ListReviewsRequest` to filter by `google_places_id`; `UpdateReview` simplified; removed tag RPCs (AddTag, RemoveTag, ListTags) — using free-form strings instead.

---

## Task 3: Regenerate protobuf code

**Step 1: Run codegen**

```bash
cd /home/gooral/Projects/resto-rate && nx run protos:generate
```

Expected: Go files regenerated in `apps/api/src/generated/reviews/v1/`, TypeScript files regenerated in `apps/web/src/lib/client/generated/reviews/v1/`. No errors.

**Step 2: Verify Go generated code compiles (it will fail — reviews service not implemented yet, that's expected)**

```bash
cd /home/gooral/Projects/resto-rate/apps/api && go build ./src/generated/...
```

Expected: PASS (generated code alone compiles fine).

---

## Task 4: Create Review GORM model

**Files:**
- Create: `apps/api/src/internal/models/review_model.go`

**Step 1: Write the file**

```go
package models

import (
	reviewspb "api/src/generated/reviews/v1"
	"time"

	"gorm.io/gorm"
)

type Review struct {
	UUIDv7
	RestaurantID   string    `gorm:"not null;index;uniqueIndex:idx_review_restaurant_user"`
	UserID         string    `gorm:"not null;index;uniqueIndex:idx_review_restaurant_user"`
	GooglePlacesID string    `gorm:"index"`
	Comment        string
	Rating         float64   `gorm:"not null"`
	Tags           []string  `gorm:"serializer:json"`
	CreatedAt      time.Time `gorm:"autoCreateTime"`
	UpdatedAt      time.Time `gorm:"autoUpdateTime"`
}

func (r *Review) BeforeCreate(tx *gorm.DB) (err error) {
	return r.UUIDv7.BeforeCreate(tx)
}

func (r *Review) ToProto() *reviewspb.ReviewProto {
	tags := r.Tags
	if tags == nil {
		tags = []string{}
	}
	return &reviewspb.ReviewProto{
		Id:             r.ID,
		UserId:         r.UserID,
		RestaurantId:   r.RestaurantID,
		GooglePlacesId: r.GooglePlacesID,
		Comment:        r.Comment,
		Rating:         r.Rating,
		Tags:           tags,
		CreatedAt:      r.CreatedAt.Unix(),
		UpdatedAt:      r.UpdatedAt.Unix(),
	}
}
```

Key design decisions:
- `uniqueIndex:idx_review_restaurant_user` — one review per user per restaurant
- `Tags []string` with `gorm:"serializer:json"` — stores as JSON text, no extra table or lib/pq needed
- `GooglePlacesID` stored directly for efficient `ListReviews` queries

---

## Task 5: Register Review model in CreateSchema

**Files:**
- Modify: `apps/api/src/internal/utils/database.go`

**Step 1: Add `Review` to the `AutoMigrate` calls**

In `CreateSchema`, after the `AutoMigrate(&models.User{})` line, add:

```go
if err := db.AutoMigrate(&models.Review{}); err != nil {
    return err
}
```

The function after the change:
```go
func CreateSchema(db *gorm.DB) error {
	slog.Info("Creating database schema...")

	if err := db.AutoMigrate(&models.Restaurant{}); err != nil {
		return err
	}

	if err := db.AutoMigrate(&models.User{}); err != nil {
		return err
	}

	if err := db.AutoMigrate(&models.Review{}); err != nil {
		return err
	}

	slog.Info("Database schema created successfully")
	return nil
}
```

---

## Task 6: Implement ReviewsService

**Files:**
- Create: `apps/api/src/services/reviews_service.go`

**Step 1: Write the file**

```go
package services

import (
	restaurantspb "api/src/generated/restaurants/v1"
	v1 "api/src/generated/reviews/v1"
	"api/src/generated/reviews/v1/v1connect"
	"api/src/internal/models"
	"context"
	"errors"
	"net/http"

	"connectrpc.com/connect"
	"github.com/valkey-io/valkey-go"
	"gorm.io/gorm"
)

type ReviewsService struct {
	v1connect.UnimplementedReviewsServiceHandler
	DB     *gorm.DB
	Valkey valkey.Client
}

func NewReviewsService(db *gorm.DB, kv valkey.Client) *ReviewsService {
	return &ReviewsService{DB: db, Valkey: kv}
}

func (s *ReviewsService) getUserIDFromSession(ctx context.Context, h http.Header) (string, error) {
	token := sessionToken(h)
	if token == "" {
		return "", connect.NewError(connect.CodeUnauthenticated, errors.New("authentication required"))
	}
	result := s.Valkey.Do(ctx, s.Valkey.B().Get().Key("session:"+token).Build())
	if result.Error() != nil {
		return "", connect.NewError(connect.CodeUnauthenticated, errors.New("session expired"))
	}
	userID, err := result.ToString()
	if err != nil {
		return "", connect.NewError(connect.CodeUnauthenticated, errors.New("invalid session"))
	}
	return userID, nil
}

func (s *ReviewsService) CreateReview(
	ctx context.Context,
	req *connect.Request[v1.CreateReviewRequest],
) (*connect.Response[v1.CreateReviewResponse], error) {
	userID, err := s.getUserIDFromSession(ctx, req.Header())
	if err != nil {
		return nil, err
	}

	if req.Msg.GooglePlacesId == "" {
		return nil, connect.NewError(connect.CodeInvalidArgument, errors.New("google_places_id is required"))
	}
	if req.Msg.Rating < 1 || req.Msg.Rating > 5 {
		return nil, connect.NewError(connect.CodeInvalidArgument, errors.New("rating must be between 1 and 5"))
	}

	var restaurant models.Restaurant
	var review models.Review

	txErr := s.DB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// Find or create restaurant by Google Places ID
		result := tx.Where(models.Restaurant{GoogleID: req.Msg.GooglePlacesId}).
			Attrs(models.Restaurant{
				Name:    req.Msg.RestaurantName,
				Address: req.Msg.RestaurantAddress,
			}).
			FirstOrCreate(&restaurant)
		if result.Error != nil {
			return result.Error
		}

		// Check for duplicate review
		var existing models.Review
		if err := tx.Where("restaurant_id = ? AND user_id = ?", restaurant.ID, userID).First(&existing).Error; err == nil {
			return connect.NewError(connect.CodeAlreadyExists, errors.New("you already reviewed this restaurant — use UpdateReview"))
		} else if !errors.Is(err, gorm.ErrRecordNotFound) {
			return err
		}

		review = models.Review{
			RestaurantID:   restaurant.ID,
			UserID:         userID,
			GooglePlacesID: req.Msg.GooglePlacesId,
			Comment:        req.Msg.Comment,
			Rating:         req.Msg.Rating,
			Tags:           req.Msg.Tags,
		}
		return tx.Create(&review).Error
	})
	if txErr != nil {
		return nil, txErr
	}

	return connect.NewResponse(&v1.CreateReviewResponse{
		Review:     review.ToProto(),
		Restaurant: restaurant.ToProto(),
	}), nil
}

func (s *ReviewsService) ListReviews(
	ctx context.Context,
	req *connect.Request[v1.ListReviewsRequest],
) (*connect.Response[v1.ListReviewsResponse], error) {
	userID, err := s.getUserIDFromSession(ctx, req.Header())
	if err != nil {
		return nil, err
	}

	var reviews []models.Review
	query := s.DB.WithContext(ctx).Where("user_id = ?", userID)
	if req.Msg.GooglePlacesId != "" {
		query = query.Where("google_places_id = ?", req.Msg.GooglePlacesId)
	}
	if err := query.Find(&reviews).Error; err != nil {
		return nil, err
	}

	protos := make([]*v1.ReviewProto, len(reviews))
	for i, r := range reviews {
		protos[i] = r.ToProto()
	}

	return connect.NewResponse(&v1.ListReviewsResponse{Reviews: protos}), nil
}

func (s *ReviewsService) UpdateReview(
	ctx context.Context,
	req *connect.Request[v1.UpdateReviewRequest],
) (*connect.Response[v1.UpdateReviewResponse], error) {
	userID, err := s.getUserIDFromSession(ctx, req.Header())
	if err != nil {
		return nil, err
	}

	if req.Msg.Rating < 1 || req.Msg.Rating > 5 {
		return nil, connect.NewError(connect.CodeInvalidArgument, errors.New("rating must be between 1 and 5"))
	}

	var review models.Review
	if err := s.DB.WithContext(ctx).First(&review, "id = ? AND user_id = ?", req.Msg.Id, userID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, connect.NewError(connect.CodeNotFound, errors.New("review not found"))
		}
		return nil, err
	}

	review.Comment = req.Msg.Comment
	review.Rating = req.Msg.Rating
	review.Tags = req.Msg.Tags

	if err := s.DB.WithContext(ctx).Save(&review).Error; err != nil {
		return nil, err
	}

	return connect.NewResponse(&v1.UpdateReviewResponse{Review: review.ToProto()}), nil
}

func (s *ReviewsService) GetReview(
	ctx context.Context,
	req *connect.Request[v1.GetReviewRequest],
) (*connect.Response[v1.GetReviewResponse], error) {
	userID, err := s.getUserIDFromSession(ctx, req.Header())
	if err != nil {
		return nil, err
	}

	var review models.Review
	if err := s.DB.WithContext(ctx).First(&review, "id = ? AND user_id = ?", req.Msg.Id, userID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, connect.NewError(connect.CodeNotFound, errors.New("review not found"))
		}
		return nil, err
	}

	return connect.NewResponse(&v1.GetReviewResponse{Review: review.ToProto()}), nil
}

func (s *ReviewsService) DeleteReview(
	ctx context.Context,
	req *connect.Request[v1.DeleteReviewRequest],
) (*connect.Response[v1.DeleteReviewResponse], error) {
	userID, err := s.getUserIDFromSession(ctx, req.Header())
	if err != nil {
		return nil, err
	}

	result := s.DB.WithContext(ctx).Where("id = ? AND user_id = ?", req.Msg.Id, userID).Delete(&models.Review{})
	if result.Error != nil {
		return nil, result.Error
	}
	if result.RowsAffected == 0 {
		return nil, connect.NewError(connect.CodeNotFound, errors.New("review not found"))
	}

	return connect.NewResponse(&v1.DeleteReviewResponse{Success: true}), nil
}

// Ensure RestaurantProto import is used
var _ = &restaurantspb.RestaurantProto{}
```

Note: `sessionToken(h)` is defined in `auth_service.go` in the same `services` package — no import needed.

---

## Task 7: Register ReviewsService in main.go

**Files:**
- Modify: `apps/api/src/main.go`

**Step 1: Add import for reviews v1connect**

In the import block, add:
```go
reviewsv1connect "api/src/generated/reviews/v1/v1connect"
```

**Step 2: Add service registration in `initializeServiceHandlers`**

Append to the returned slice:
```go
func() ServiceRegistration {
    svc := services.NewReviewsService(db, valkeyClient)
    path, handler := reviewsv1connect.NewReviewsServiceHandler(svc, connect.WithInterceptors(prometheusInterceptor))
    return ServiceRegistration{Path: path, Handler: handler}
}(),
```

**Step 3: Add to gRPC reflection in `optionallySetupGRPCReflection`**

Add `reviewsv1connect.ReviewsServiceName` to the `grpcreflect.NewStaticReflector(...)` call.

**Step 4: Verify Go compiles**

```bash
cd /home/gooral/Projects/resto-rate/apps/api && go build ./...
```

Expected: PASS with no errors.

**Step 5: Commit backend**

```bash
git add apps/api/ packages/protos/reviews/
git commit -m "feat: implement ReviewsService with atomic create-or-find restaurant

Co-Authored-By: Claude Sonnet 4.6 <noreply@anthropic.com>"
```

---

## Task 8: Add reviews client to client.ts

**Files:**
- Modify: `apps/web/src/lib/client/client.ts`

**Step 1: Replace the entire file**

```typescript
import { createConnectTransport } from '@connectrpc/connect-web';
import { createClient } from '@connectrpc/connect';
import { RestaurantsService } from '$lib/client/generated/restaurants/v1/restaurants_service_pb';
import { UsersService } from '$lib/client/generated/users/v1/users_service_pb';
import { GoogleMapsService } from '$lib/client/generated/google_maps/v1/google_maps_service_pb';
import { AuthService } from '$lib/client/generated/auth/v1/auth_service_pb';
import { ReviewsService } from '$lib/client/generated/reviews/v1/reviews_service_pb';

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

export default { restaurants, users, googleMaps, auth, reviews };
```

> **Note:** After proto regeneration the `ReviewsService` is exported from `reviews_service_pb.ts`. If the import fails, check `apps/web/src/lib/client/generated/reviews/v1/reviews_service_pb.ts` — the service descriptor may be in `reviews_service_connect.ts` instead (use the connect file in that case).

---

## Task 9: Create PlacePreviewCard.svelte

**Files:**
- Create: `apps/web/src/lib/ui/components/PlacePreviewCard.svelte`

**Step 1: Write the file**

```svelte
<script lang="ts">
	import type { Place } from '$lib/client/generated/google_maps/v1/google_maps_service_pb';
	import { MapPin, Star } from '@lucide/svelte';

	const { place } = $props<{ place: Place }>();

	const name = $derived(place.displayName?.text || place.name || '');
	const address = $derived(place.formattedAddress || '');
	const rating = $derived(place.rating ?? null);
	const reviewCount = $derived(place.userRatingCount ?? null);
</script>

<div class="rounded-2xl bg-white p-6 shadow-xl">
	<div class="mb-2">
		<span class="rounded-full bg-blue-100 px-2 py-0.5 text-xs font-medium text-blue-700">
			Preview — not saved yet
		</span>
	</div>

	<h3 class="text-xl font-bold text-gray-900">{name}</h3>

	{#if address}
		<div class="mt-2 flex items-start gap-2 text-sm text-gray-600">
			<MapPin class="mt-0.5 h-4 w-4 shrink-0 text-gray-400" />
			<span>{address}</span>
		</div>
	{/if}

	{#if rating !== null}
		<div class="mt-2 flex items-center gap-1.5">
			{#each Array(5) as _, i}
				<Star
					class="h-4 w-4 {i < Math.round(rating)
						? 'fill-amber-400 text-amber-400'
						: 'fill-none text-gray-300'}"
				/>
			{/each}
			<span class="text-sm text-gray-600">{rating.toFixed(1)}</span>
			{#if reviewCount}
				<span class="text-xs text-gray-400">({reviewCount.toLocaleString()} Google reviews)</span>
			{/if}
		</div>
	{/if}
</div>
```

---

## Task 10: Create RatingForm.svelte

**Files:**
- Create: `apps/web/src/lib/ui/components/RatingForm.svelte`

**Step 1: Write the file**

```svelte
<script lang="ts">
	import { Button } from '$lib/components/ui/button/index.js';
	import { Label } from '$lib/components/ui/label/index.js';
	import { Star, X } from '@lucide/svelte';
	import client from '$lib/client/client';
	import type { ReviewProto } from '$lib/client/generated/reviews/v1/review_pb';

	const {
		googlePlacesId,
		restaurantName,
		restaurantAddress,
		existingReview,
		onSubmit
	} = $props<{
		googlePlacesId: string;
		restaurantName: string;
		restaurantAddress: string;
		existingReview?: ReviewProto;
		onSubmit: (review: ReviewProto) => void;
	}>();

	let rating = $state(existingReview?.rating ?? 0);
	let hoverRating = $state(0);
	let comment = $state(existingReview?.comment ?? '');
	let tags = $state<string[]>(existingReview?.tags ? [...existingReview.tags] : []);
	let tagInput = $state('');
	let loading = $state(false);
	let error = $state<string | null>(null);

	const isEdit = $derived(!!existingReview?.id);
	const displayRating = $derived(hoverRating || rating);

	function addTag() {
		const t = tagInput.trim().replace(/,$/, '');
		if (t && !tags.includes(t)) {
			tags = [...tags, t];
		}
		tagInput = '';
	}

	function removeTag(tag: string) {
		tags = tags.filter((t) => t !== tag);
	}

	function handleTagKeydown(e: KeyboardEvent) {
		if (e.key === 'Enter' || e.key === ',') {
			e.preventDefault();
			addTag();
		}
	}

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
		<Label for="tag-input" class="mb-1 block text-sm">Tags (optional)</Label>
		<div class="flex flex-wrap gap-1.5 mb-2">
			{#each tags as tag}
				<span class="flex items-center gap-1 rounded-full bg-blue-100 px-2.5 py-0.5 text-xs font-medium text-blue-700">
					{tag}
					<button
						type="button"
						onclick={() => removeTag(tag)}
						class="text-blue-500 hover:text-blue-700"
						aria-label="Remove tag {tag}"
					>
						<X class="h-3 w-3" />
					</button>
				</span>
			{/each}
		</div>
		<input
			id="tag-input"
			type="text"
			bind:value={tagInput}
			onkeydown={handleTagKeydown}
			onblur={addTag}
			placeholder="Type a tag and press Enter"
			class="w-full rounded-lg border border-gray-300 px-3 py-1.5 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500"
		/>
		<p class="mt-1 text-xs text-gray-400">Press Enter or comma to add a tag</p>
	</div>

	{#if error}
		<p class="mb-3 text-sm text-red-600">{error}</p>
	{/if}

	<Button onclick={handleSubmit} disabled={loading || rating < 1} class="w-full">
		{loading ? 'Saving…' : isEdit ? 'Update rating' : 'Save rating'}
	</Button>
</div>
```

---

## Task 11: Create ReviewSummary.svelte

**Files:**
- Create: `apps/web/src/lib/ui/components/ReviewSummary.svelte`

**Step 1: Write the file**

```svelte
<script lang="ts">
	import { Button } from '$lib/components/ui/button/index.js';
	import { Star } from '@lucide/svelte';
	import type { ReviewProto } from '$lib/client/generated/reviews/v1/review_pb';

	const { review, onEdit } = $props<{
		review: ReviewProto;
		onEdit: () => void;
	}>();
</script>

<div class="rounded-2xl bg-white p-6 shadow-xl">
	<div class="mb-3 flex items-center justify-between">
		<h4 class="text-base font-semibold text-gray-800">Your rating</h4>
		<Button size="sm" variant="outline" onclick={onEdit}>Edit</Button>
	</div>

	<!-- Stars (read-only) -->
	<div class="mb-3 flex items-center gap-0.5">
		{#each Array(5) as _, i}
			<Star
				class="h-5 w-5 {i < review.rating
					? 'fill-amber-400 text-amber-400'
					: 'fill-none text-gray-300'}"
			/>
		{/each}
		<span class="ml-2 text-sm font-semibold text-gray-800">{review.rating.toFixed(1)}</span>
	</div>

	{#if review.comment}
		<p class="mb-3 text-sm leading-relaxed text-gray-600">{review.comment}</p>
	{/if}

	{#if review.tags && review.tags.length > 0}
		<div class="flex flex-wrap gap-1.5">
			{#each review.tags as tag}
				<span class="rounded-full bg-blue-100 px-2.5 py-0.5 text-xs font-medium text-blue-700">
					{tag}
				</span>
			{/each}
		</div>
	{/if}
</div>
```

---

## Task 12: Rewrite RestaurantSearchSv.svelte

**Files:**
- Modify: `apps/web/src/lib/ui/components/RestaurantSearchSv.svelte`

**Step 1: Replace the entire file**

```svelte
<script lang="ts">
	import clients from '$lib/client/client';
	import type {
		Place,
		Suggestion
	} from '$lib/client/generated/google_maps/v1/google_maps_service_pb';
	import type { ReviewProto } from '$lib/client/generated/reviews/v1/review_pb';
	import { onMount, onDestroy } from 'svelte';
	import { Input } from '$lib/components/ui/input/index.js';
	import { v4 as uuidv4 } from 'uuid';
	import { auth } from '$lib/state/auth.svelte';
	import PlacePreviewCard from './PlacePreviewCard.svelte';
	import RatingForm from './RatingForm.svelte';
	import ReviewSummary from './ReviewSummary.svelte';

	function randomUUID(): string {
		return uuidv4();
	}

	let autocompleteSessionToken = randomUUID();
	let input = $state('');
	let suggestions = $state<Suggestion[]>([]);
	let isLoading = $state(false);
	let selectedIndex = $state(-1);
	let showSuggestions = $state(false);
	let queryPrediction = $state('');

	let selectedPlace = $state<Place | null>(null);
	let isCheckingReview = $state(false);
	let currentReview = $state<ReviewProto | null>(null);
	let isEditingReview = $state(false);

	let debounceTimer: ReturnType<typeof setTimeout> | null = null;

	let { onPlaceSelected } = $props<{
		onPlaceSelected?: (place: Place) => void;
	}>();

	function debouncedAutocomplete(input: string) {
		if (debounceTimer) clearTimeout(debounceTimer);

		if (input.length < 2) {
			suggestions = [];
			showSuggestions = false;
			queryPrediction = '';
			autocompleteSessionToken = randomUUID();
			return;
		}

		if (input.length >= 2 && !isLoading) isLoading = true;

		debounceTimer = setTimeout(() => {
			if (input.length >= 2) performAutocomplete(input);
		}, 300);
	}

	async function performAutocomplete(input: string, regionCode: string = 'pl') {
		if (input.length < 2) return;
		isLoading = true;
		try {
			const response = await clients.googleMaps.autocompletePlaces({
				input,
				languageCode: 'pl',
				includedRegionCodes: [regionCode],
				sessionToken: autocompleteSessionToken,
				includeQueryPrediction: true
			});

			suggestions = response.suggestions || [];
			showSuggestions = suggestions.length > 0;

			const querySuggestion = suggestions.find((s) => s.queryPrediction);
			queryPrediction = querySuggestion?.queryPrediction?.text?.text ?? '';
			selectedIndex = -1;
		} catch (error) {
			console.error('Autocomplete error:', error);
			suggestions = [];
			showSuggestions = false;
		} finally {
			isLoading = false;
		}
	}

	async function getPlaceDetails(name: string) {
		try {
			const place = await clients.googleMaps.getRestaurantDetails({
				name,
				languageCode: 'pl',
				regionCode: 'pl',
				sessionToken: autocompleteSessionToken
			});

			selectedPlace = place;
			currentReview = null;
			isEditingReview = false;
			suggestions = [];
			showSuggestions = false;
			queryPrediction = '';
			input = place.displayName?.text || place.name || '';
			autocompleteSessionToken = randomUUID();

			if (onPlaceSelected) onPlaceSelected(place);

			// Check if current user already has a review for this place
			isCheckingReview = true;
			try {
				const res = await clients.reviews.listReviews({ googlePlacesId: place.name || '' });
				currentReview = res.reviews?.[0] ?? null;
			} catch {
				currentReview = null;
			} finally {
				isCheckingReview = false;
			}
		} catch (error) {
			console.error('Get place details error:', error);
		}
	}

	function handleInputChange(event: Event) {
		const target = event.target as HTMLInputElement;
		input = target.value;

		if (!input.trim()) {
			suggestions = [];
			showSuggestions = false;
			queryPrediction = '';
			isLoading = false;
			return;
		}

		debouncedAutocomplete(input);
	}

	function handleKeyDown(event: KeyboardEvent) {
		if (!showSuggestions) return;
		switch (event.key) {
			case 'ArrowDown':
				event.preventDefault();
				selectedIndex = Math.min(selectedIndex + 1, suggestions.length - 1);
				break;
			case 'ArrowUp':
				event.preventDefault();
				selectedIndex = Math.max(selectedIndex - 1, -1);
				break;
			case 'Enter':
				event.preventDefault();
				if (selectedIndex >= 0 && selectedIndex < suggestions.length) {
					selectSuggestion(suggestions[selectedIndex]);
				}
				break;
			case 'Escape':
				showSuggestions = false;
				selectedIndex = -1;
				break;
		}
	}

	function selectSuggestion(suggestion: Suggestion) {
		if (suggestion.placePrediction?.place) {
			getPlaceDetails(suggestion.placePrediction.place);
		}
	}

	function getSuggestionText(suggestion: Suggestion): string {
		return (
			suggestion.placePrediction?.structuredFormat?.mainText?.text ||
			suggestion.placePrediction?.text?.text ||
			''
		);
	}

	function getSuggestionSubtext(suggestion: Suggestion): string {
		return suggestion.placePrediction?.structuredFormat?.secondaryText?.text || '';
	}

	onMount(() => {
		autocompleteSessionToken = randomUUID();
	});

	onDestroy(() => {
		if (debounceTimer) {
			clearTimeout(debounceTimer);
			debounceTimer = null;
		}
	});
</script>

{#if !auth.isLoggedIn}
	<p class="text-sm text-gray-500">Please log in to search restaurants.</p>
{:else}
	<div class="relative w-full max-w-md">
		<div class="relative flex items-center">
			<Input
				type="text"
				bind:value={input}
				oninput={handleInputChange}
				onkeydown={handleKeyDown}
				placeholder="Search for restaurants..."
				class="w-full bg-[url('/GoogleMaps_Logo_Gray.svg')] bg-[length:60px_60px] bg-[position:calc(100%-2.25rem)_50%] bg-no-repeat pr-10"
			/>
			{#if isLoading}
				<div class="absolute right-3 flex items-center">
					<div class="border-t-primary-500 h-4 w-4 animate-spin rounded-full border-2 border-gray-300"></div>
				</div>
			{/if}
		</div>

		{#if showSuggestions && suggestions.length > 0}
			<div
				class="absolute left-0 right-0 top-full z-50 max-h-80 overflow-y-auto rounded-b-lg border-2 border-t-0 border-gray-200 bg-white shadow-lg"
			>
				{#each suggestions as suggestion, index}
					{#if suggestion.placePrediction}
						<div
							class="cursor-pointer border-b border-gray-100 p-3 transition-colors duration-200 last:border-b-0 hover:bg-gray-50 {index === selectedIndex ? 'bg-gray-50' : ''}"
							onclick={() => selectSuggestion(suggestion)}
							onkeydown={(e) => e.key === 'Enter' && selectSuggestion(suggestion)}
							onmouseenter={() => (selectedIndex = index)}
							tabindex="0"
							role="button"
							aria-label="Select {getSuggestionText(suggestion)}"
						>
							<div class="mb-1 font-medium text-gray-900">{getSuggestionText(suggestion)}</div>
							{#if getSuggestionSubtext(suggestion)}
								<div class="text-sm text-gray-500">{getSuggestionSubtext(suggestion)}</div>
							{/if}
						</div>
					{/if}
				{/each}
			</div>
		{/if}

		{#if queryPrediction && input.length > 0}
			<div
				class="pointer-events-none absolute left-4 top-1/2 z-10 -translate-y-1/2 transform text-base text-gray-500 {showSuggestions ? 'hidden' : ''}"
			>
				<span class="text-transparent">{input}</span>
				<span class="text-gray-500 opacity-60">{queryPrediction.substring(input.length)}</span>
			</div>
		{/if}

		{#if input.length > 0 && input.length < 2}
			<div class="absolute left-0 right-0 top-full mt-1 rounded border border-gray-200 bg-gray-50 p-2 text-sm text-gray-500">
				Type at least 2 characters to search...
			</div>
		{/if}
	</div>

	{#if selectedPlace}
		<div class="mt-6 space-y-4">
			<PlacePreviewCard place={selectedPlace} />

			{#if isCheckingReview}
				<div class="flex items-center gap-2 text-sm text-gray-500">
					<div class="h-4 w-4 animate-spin rounded-full border-2 border-gray-300 border-t-blue-500"></div>
					Checking your review…
				</div>
			{:else if currentReview && !isEditingReview}
				<ReviewSummary
					review={currentReview}
					onEdit={() => (isEditingReview = true)}
				/>
			{:else}
				<RatingForm
					googlePlacesId={selectedPlace.name || ''}
					restaurantName={selectedPlace.displayName?.text || selectedPlace.name || ''}
					restaurantAddress={selectedPlace.formattedAddress || ''}
					existingReview={isEditingReview ? currentReview ?? undefined : undefined}
					onSubmit={(review) => {
						currentReview = review;
						isEditingReview = false;
					}}
				/>
			{/if}
		</div>
	{/if}
{/if}
```

---

## Task 13: Type-check and commit

**Step 1: Run Svelte type checker**

```bash
cd /home/gooral/Projects/resto-rate/apps/web && bun run check
```

Expected: `svelte-check found 0 errors and 0 warnings`.

If there are errors about `ReviewsService` import path, check what's in:
```bash
head -20 apps/web/src/lib/client/generated/reviews/v1/reviews_service_pb.ts
```
If `ReviewsService` is not exported there, check `reviews_service_connect.ts` and update the import in `client.ts` accordingly.

**Step 2: Verify no flowbite references**

```bash
grep -r "flowbite\|createRestaurant" apps/web/src/ || echo "clean"
```

Expected: `clean` (or only false positives in non-source files).

**Step 3: Verify Go builds**

```bash
cd /home/gooral/Projects/resto-rate/apps/api && go build ./...
```

Expected: PASS.

**Step 4: Commit**

```bash
git add apps/web/src/ apps/web/src/lib/client/client.ts
git commit -m "feat: deferred restaurant save + rating form

- Remove createRestaurant call from search flow
- Auth gate: show login prompt when not authenticated
- PlacePreviewCard: preview-only Google Places card (unsaved state)
- RatingForm: star picker + comment + free-form tag chips
- ReviewSummary: read-only review display with Edit button
- ListReviews on place select to pre-populate existing review

Co-Authored-By: Claude Sonnet 4.6 <noreply@anthropic.com>"
```

---

## Troubleshooting

**`ReviewsService` not found in generated TypeScript:**
Run `nx run protos:generate` first. If it's in `reviews_service_connect.ts` instead of `reviews_service_pb.ts`, update the import in `client.ts`:
```typescript
import { ReviewsService } from '$lib/client/generated/reviews/v1/reviews_service_connect';
```

**`gorm:"serializer:json"` not working on `[]string`:**
GORM v1.23+ supports the JSON serializer. The project uses `gorm.io/gorm v1.30.0` so this works. If you see `unsupported data type: []string`, add `gorm:"type:text;serializer:json"` to be explicit.

**`sessionToken` function not found in reviews_service.go:**
It's defined in `auth_service.go` in the same `services` package — it's package-level, not a method, so it's visible. No import needed.

**`FirstOrCreate` not creating with correct fields:**
The `Attrs()` pattern sets fields only when creating, not when finding. If you see empty `Name`/`Address` on existing records that's expected — existing records keep their data.

**Tags showing as `null` instead of `[]`:**
The `ToProto()` method handles `nil` slice by initializing to `[]string{}`. If the client still gets null, check the protobuf JSON serialization — `repeated string` with no elements serializes to `[]` in proto3.
