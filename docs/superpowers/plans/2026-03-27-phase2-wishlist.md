# Phase 2 PR2: Wishlist Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add a Wishlist feature — users can save restaurants to a wishlist, and creating a review auto-removes the restaurant from their wishlist.

**Architecture:** New `WishlistService` (Add, Remove, List) backed by a `wishlist_items` table with a unique `(user_id, restaurant_id)` constraint. The existing `CreateReview` transaction gains a wishlist-deletion step. The home page search flow gains Add/Remove wishlist buttons alongside the existing Rate button.

**Tech Stack:** Go + Connect-RPC + GORM (backend), Protocol Buffers (API contract), SvelteKit 5 + ShadCN Svelte (frontend), Bun + Nx (tooling)

---

## File Structure

**Create:**
- `packages/protos/wishlist/v1/wishlist_item.proto` — WishlistItemProto message
- `packages/protos/wishlist/v1/wishlist_service.proto` — WishlistService RPCs
- `apps/api/src/internal/models/wishlist_item_model.go` — GORM model + ToProto()
- `apps/api/src/services/wishlist_service.go` — Add/Remove/List handlers
- `apps/api/src/test/wishlist_service_test.go` — unit tests (nil DB guards)

**Modify:**
- `apps/api/src/internal/utils/database.go` — add WishlistItem to CreateSchema
- `apps/api/src/services/reviews_service.go` — delete wishlist item in CreateReview transaction
- `apps/api/src/main.go` — import wishlistv1connect, register service, add to gRPC reflection
- `apps/web/src/lib/client/client.ts` — add wishlist client
- `apps/web/src/lib/ui/components/RestaurantSearchSv.svelte` — wishlist state + buttons

---

## Task 1: Proto files + codegen

**Files:**
- Create: `packages/protos/wishlist/v1/wishlist_item.proto`
- Create: `packages/protos/wishlist/v1/wishlist_service.proto`

- [ ] **Step 1: Create the wishlist item proto**

```
packages/protos/wishlist/v1/wishlist_item.proto
```

```protobuf
syntax = "proto3";

package wishlist.v1;

option go_package = "api/src/generated/wishlist/v1";

message WishlistItemProto {
  string id = 1;
  string google_places_id = 2;
  string restaurant_name = 3;
  string restaurant_address = 4;
  string city = 5;
  string country = 6;
  int64 created_at = 7;
}
```

- [ ] **Step 2: Create the wishlist service proto**

```
packages/protos/wishlist/v1/wishlist_service.proto
```

```protobuf
syntax = "proto3";

package wishlist.v1;

import "wishlist/v1/wishlist_item.proto";

option go_package = "api/src/generated/wishlist/v1";

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
}

message ListWishlistResponse {
  repeated WishlistItemProto items = 1;
}
```

- [ ] **Step 3: Generate Go + TypeScript**

Run from the workspace root:
```bash
bunx nx run protos:generate:api
bunx nx run protos:generate:web
```

Expected: new directories `apps/api/src/generated/wishlist/v1/` and `apps/web/src/lib/client/generated/wishlist/v1/` appear. No errors.

- [ ] **Step 4: Verify Go compiles with new generated code**

```bash
bunx nx run api:build
```

Expected: `Successfully ran target build for project api`

- [ ] **Step 5: Commit**

```bash
git add packages/protos/wishlist/ apps/api/src/generated/ apps/web/src/lib/client/generated/
git commit -m "feat: wishlist proto + codegen"
```

Note: the generated Go and TS files are gitignored — skip them. Only commit the proto sources:
```bash
git add packages/protos/wishlist/
git commit -m "feat: wishlist proto"
```

---

## Task 2: WishlistItem model + schema

**Files:**
- Create: `apps/api/src/internal/models/wishlist_item_model.go`
- Modify: `apps/api/src/internal/utils/database.go`

- [ ] **Step 1: Write the failing test**

Create `apps/api/src/test/wishlist_service_test.go`:

```go
package test

import (
	"api/src/services"
	wishlistv1 "api/src/generated/wishlist/v1"
	"context"
	"testing"

	"connectrpc.com/connect"
)

func TestWishlistService_AddToWishlist_NilDB(t *testing.T) {
	svc := &services.WishlistService{}
	req := connect.NewRequest(&wishlistv1.AddToWishlistRequest{
		GooglePlacesId: "places/abc123",
	})
	_, err := svc.AddToWishlist(context.Background(), req)
	if err == nil {
		t.Fatal("expected error from nil DB, got nil")
	}
}

func TestWishlistService_RemoveFromWishlist_NilDB(t *testing.T) {
	svc := &services.WishlistService{}
	req := connect.NewRequest(&wishlistv1.RemoveFromWishlistRequest{
		GooglePlacesId: "places/abc123",
	})
	_, err := svc.RemoveFromWishlist(context.Background(), req)
	if err == nil {
		t.Fatal("expected error from nil DB, got nil")
	}
}

func TestWishlistService_ListWishlist_NilDB(t *testing.T) {
	svc := &services.WishlistService{}
	req := connect.NewRequest(&wishlistv1.ListWishlistRequest{})
	_, err := svc.ListWishlist(context.Background(), req)
	if err == nil {
		t.Fatal("expected error from nil DB, got nil")
	}
}
```

- [ ] **Step 2: Run the test — confirm it fails to compile**

```bash
cd apps/api && go test ./src/test/... -run TestWishlistService -v 2>&1 | head -20
```

Expected: compile error — `services.WishlistService` undefined.

- [ ] **Step 3: Create the WishlistItem model**

Create `apps/api/src/internal/models/wishlist_item_model.go`:

```go
package models

import (
	wishlistv1 "api/src/generated/wishlist/v1"
	"time"

	"gorm.io/gorm"
)

type WishlistItem struct {
	UUIDv7
	UserID         string     `gorm:"not null;index;uniqueIndex:idx_wishlist_user_restaurant"`
	RestaurantID   string     `gorm:"not null;index;uniqueIndex:idx_wishlist_user_restaurant"`
	Restaurant     Restaurant `gorm:"foreignKey:RestaurantID"`
	GooglePlacesID string     `gorm:"not null;index"`
	CreatedAt      time.Time  `gorm:"autoCreateTime"`
}

func (w *WishlistItem) BeforeCreate(tx *gorm.DB) (err error) {
	return w.UUIDv7.BeforeCreate(tx)
}

func (w *WishlistItem) ToProto() *wishlistv1.WishlistItemProto {
	return &wishlistv1.WishlistItemProto{
		Id:                w.ID,
		GooglePlacesId:    w.GooglePlacesID,
		RestaurantName:    w.Restaurant.Name,
		RestaurantAddress: w.Restaurant.Address,
		City:              w.Restaurant.City,
		Country:           w.Restaurant.Country,
		CreatedAt:         w.CreatedAt.UnixMilli(),
	}
}
```

- [ ] **Step 4: Add WishlistItem to CreateSchema**

In `apps/api/src/internal/utils/database.go`, find the `CreateSchema` function and add `&models.WishlistItem{}` to the `AutoMigrate` call.

Current call looks like:
```go
err := db.AutoMigrate(
    &models.User{},
    &models.Restaurant{},
    &models.Review{},
    &models.Tag{},
)
```

Add `&models.WishlistItem{}`:
```go
err := db.AutoMigrate(
    &models.User{},
    &models.Restaurant{},
    &models.Review{},
    &models.Tag{},
    &models.WishlistItem{},
)
```

- [ ] **Step 5: Create a stub WishlistService so the test compiles**

Create `apps/api/src/services/wishlist_service.go` with just enough to compile:

```go
package services

import (
	wishlistv1 "api/src/generated/wishlist/v1"
	"api/src/generated/wishlist/v1/v1connect"
	"context"
	"errors"

	"connectrpc.com/connect"
	"github.com/valkey-io/valkey-go"
	"gorm.io/gorm"
)

type WishlistService struct {
	v1connect.UnimplementedWishlistServiceHandler
	DB     *gorm.DB
	Valkey valkey.Client
}

func NewWishlistService(db *gorm.DB, kv valkey.Client) *WishlistService {
	return &WishlistService{DB: db, Valkey: kv}
}

func (s *WishlistService) AddToWishlist(
	ctx context.Context,
	req *connect.Request[wishlistv1.AddToWishlistRequest],
) (*connect.Response[wishlistv1.AddToWishlistResponse], error) {
	if s.DB == nil {
		return nil, connect.NewError(connect.CodeInternal, errors.New("database not initialized"))
	}
	return nil, connect.NewError(connect.CodeUnimplemented, errors.New("not implemented"))
}

func (s *WishlistService) RemoveFromWishlist(
	ctx context.Context,
	req *connect.Request[wishlistv1.RemoveFromWishlistRequest],
) (*connect.Response[wishlistv1.RemoveFromWishlistResponse], error) {
	if s.DB == nil {
		return nil, connect.NewError(connect.CodeInternal, errors.New("database not initialized"))
	}
	return nil, connect.NewError(connect.CodeUnimplemented, errors.New("not implemented"))
}

func (s *WishlistService) ListWishlist(
	ctx context.Context,
	req *connect.Request[wishlistv1.ListWishlistRequest],
) (*connect.Response[wishlistv1.ListWishlistResponse], error) {
	if s.DB == nil {
		return nil, connect.NewError(connect.CodeInternal, errors.New("database not initialized"))
	}
	return nil, connect.NewError(connect.CodeUnimplemented, errors.New("not implemented"))
}
```

- [ ] **Step 6: Run the tests — confirm they pass**

```bash
cd apps/api && go test ./src/test/... -run TestWishlistService -v
```

Expected:
```
--- PASS: TestWishlistService_AddToWishlist_NilDB (0.00s)
--- PASS: TestWishlistService_RemoveFromWishlist_NilDB (0.00s)
--- PASS: TestWishlistService_ListWishlist_NilDB (0.00s)
```

- [ ] **Step 7: Verify full build**

```bash
bunx nx run api:build
```

Expected: no errors.

- [ ] **Step 8: Commit**

```bash
git add apps/api/src/internal/models/wishlist_item_model.go \
        apps/api/src/internal/utils/database.go \
        apps/api/src/services/wishlist_service.go \
        apps/api/src/test/wishlist_service_test.go
git commit -m "feat: WishlistItem model + WishlistService stub + tests"
```

---

## Task 3: Implement WishlistService handlers

**Files:**
- Modify: `apps/api/src/services/wishlist_service.go`

- [ ] **Step 1: Implement AddToWishlist**

Replace the `AddToWishlist` method stub in `apps/api/src/services/wishlist_service.go`:

```go
func (s *WishlistService) AddToWishlist(
	ctx context.Context,
	req *connect.Request[wishlistv1.AddToWishlistRequest],
) (*connect.Response[wishlistv1.AddToWishlistResponse], error) {
	if s.DB == nil {
		return nil, connect.NewError(connect.CodeInternal, errors.New("database not initialized"))
	}

	userID, err := getUserIDFromSession(ctx, req.Header(), s.Valkey)
	if err != nil {
		return nil, err
	}

	if req.Msg.GooglePlacesId == "" {
		return nil, connect.NewError(connect.CodeInvalidArgument, errors.New("google_places_id is required"))
	}

	var restaurant models.Restaurant
	result := s.DB.WithContext(ctx).
		Where(models.Restaurant{GoogleID: req.Msg.GooglePlacesId}).
		Attrs(models.Restaurant{
			Name:    req.Msg.RestaurantName,
			Address: req.Msg.RestaurantAddress,
			City:    req.Msg.City,
			Country: req.Msg.Country,
		}).
		FirstOrCreate(&restaurant)
	if result.Error != nil {
		return nil, result.Error
	}

	item := models.WishlistItem{
		UserID:         userID,
		RestaurantID:   restaurant.ID,
		GooglePlacesID: req.Msg.GooglePlacesId,
	}

	// Use FirstOrCreate so calling Add twice is idempotent
	var existing models.WishlistItem
	res := s.DB.WithContext(ctx).
		Where("user_id = ? AND restaurant_id = ?", userID, restaurant.ID).
		FirstOrCreate(&existing, item)
	if res.Error != nil {
		return nil, res.Error
	}

	// Preload restaurant for ToProto
	existing.Restaurant = restaurant

	return connect.NewResponse(&wishlistv1.AddToWishlistResponse{
		Item: existing.ToProto(),
	}), nil
}
```

- [ ] **Step 2: Extract session helper**

`AddToWishlist` calls `getUserIDFromSession` — this is a package-level helper that needs to be shared between `reviews_service.go` and `wishlist_service.go`. Currently it's a method on `ReviewsService`. Extract it to a package-level function.

In `apps/api/src/services/reviews_service.go`, change:
```go
func (s *ReviewsService) getUserIDFromSession(ctx context.Context, h http.Header) (string, error) {
```
to a package-level function (remove the receiver):
```go
func getUserIDFromSession(ctx context.Context, h http.Header, kv valkey.Client) (string, error) {
	token := sessionToken(h)
	if token == "" {
		return "", connect.NewError(connect.CodeUnauthenticated, errors.New("authentication required"))
	}
	result := kv.Do(ctx, kv.B().Get().Key("session:"+token).Build())
	if result.Error() != nil {
		return "", connect.NewError(connect.CodeUnauthenticated, errors.New("session expired"))
	}
	userID, err := result.ToString()
	if err != nil {
		return "", connect.NewError(connect.CodeUnauthenticated, errors.New("invalid session"))
	}
	return userID, nil
}
```

Update all call sites in `reviews_service.go` from `s.getUserIDFromSession(ctx, req.Header())` to `getUserIDFromSession(ctx, req.Header(), s.Valkey)`. There are 5 call sites — update them all:
- `CreateReview`
- `ListReviews`
- `UpdateReview`
- `GetReview`
- `DeleteReview`

Also remove the `net/http` import from `reviews_service.go` if it becomes unused — check after the refactor (`sessionToken` is in `auth_service.go` in the same package, so `net/http` may still be needed for `http.Header`).

- [ ] **Step 3: Implement RemoveFromWishlist**

Replace the `RemoveFromWishlist` stub:

```go
func (s *WishlistService) RemoveFromWishlist(
	ctx context.Context,
	req *connect.Request[wishlistv1.RemoveFromWishlistRequest],
) (*connect.Response[wishlistv1.RemoveFromWishlistResponse], error) {
	if s.DB == nil {
		return nil, connect.NewError(connect.CodeInternal, errors.New("database not initialized"))
	}

	userID, err := getUserIDFromSession(ctx, req.Header(), s.Valkey)
	if err != nil {
		return nil, err
	}

	if req.Msg.GooglePlacesId == "" {
		return nil, connect.NewError(connect.CodeInvalidArgument, errors.New("google_places_id is required"))
	}

	result := s.DB.WithContext(ctx).
		Where("user_id = ? AND google_places_id = ?", userID, req.Msg.GooglePlacesId).
		Delete(&models.WishlistItem{})
	if result.Error != nil {
		return nil, result.Error
	}

	return connect.NewResponse(&wishlistv1.RemoveFromWishlistResponse{Success: true}), nil
}
```

- [ ] **Step 4: Implement ListWishlist**

Replace the `ListWishlist` stub:

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

	query := s.DB.WithContext(ctx).
		Preload("Restaurant").
		Where("user_id = ?", userID)

	if req.Msg.GooglePlacesId != "" {
		query = query.Where("google_places_id = ?", req.Msg.GooglePlacesId)
	}

	var items []models.WishlistItem
	if err := query.Order("created_at desc").Find(&items).Error; err != nil {
		return nil, err
	}

	protos := make([]*wishlistv1.WishlistItemProto, len(items))
	for i, item := range items {
		protos[i] = item.ToProto()
	}

	return connect.NewResponse(&wishlistv1.ListWishlistResponse{Items: protos}), nil
}
```

Also add the missing imports at the top of `wishlist_service.go`. The full import block:

```go
import (
	wishlistv1 "api/src/generated/wishlist/v1"
	"api/src/generated/wishlist/v1/v1connect"
	"api/src/internal/models"
	"context"
	"errors"
	"net/http"

	"connectrpc.com/connect"
	"github.com/valkey-io/valkey-go"
	"gorm.io/gorm"
)
```

- [ ] **Step 5: Run all tests**

```bash
cd apps/api && go test ./src/test/... -v
```

Expected: all tests pass (including the 3 wishlist tests from Task 2).

- [ ] **Step 6: Verify build**

```bash
bunx nx run api:build
```

- [ ] **Step 7: Commit**

```bash
git add apps/api/src/services/wishlist_service.go \
        apps/api/src/services/reviews_service.go
git commit -m "feat: implement WishlistService + extract getUserIDFromSession"
```

---

## Task 4: Wire WishlistService into main.go

**Files:**
- Modify: `apps/api/src/main.go`

- [ ] **Step 1: Add import**

In `apps/api/src/main.go`, add to the import block:

```go
wishlistv1connect "api/src/generated/wishlist/v1/v1connect"
```

- [ ] **Step 2: Register the service**

In `initializeServiceHandlers`, add after the TagsService registration:

```go
func() ServiceRegistration {
    svc := services.NewWishlistService(db, valkeyClient)
    path, handler := wishlistv1connect.NewWishlistServiceHandler(svc, connect.WithInterceptors(prometheusInterceptor))
    return ServiceRegistration{Path: path, Handler: handler}
}(),
```

- [ ] **Step 3: Add to gRPC reflection**

In `optionallySetupGRPCReflection`, add `wishlistv1connect.WishlistServiceName` to the `grpcreflect.NewStaticReflector(...)` call:

```go
reflector := grpcreflect.NewStaticReflector(
    usersv1connect.UsersServiceName,
    restaurantsv1connect.RestaurantsServiceName,
    googlemapsv1connect.GoogleMapsServiceName,
    authv1connect.AuthServiceName,
    reviewsv1connect.ReviewsServiceName,
    tagsv1connect.TagsServiceName,
    wishlistv1connect.WishlistServiceName,
)
```

- [ ] **Step 4: Build and test**

```bash
bunx nx run api:build && cd apps/api && go test ./src/test/... -v
```

Expected: build succeeds, all tests pass.

- [ ] **Step 5: Commit**

```bash
git add apps/api/src/main.go
git commit -m "feat: register WishlistService in main.go"
```

---

## Task 5: Delete wishlist item in CreateReview transaction

**Files:**
- Modify: `apps/api/src/services/reviews_service.go`

- [ ] **Step 1: Add wishlist deletion inside the transaction**

In `reviews_service.go`, inside `CreateReview`'s `txErr := s.DB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {` block, add the wishlist deletion **after** finding/creating the restaurant and **before** creating the review:

```go
txErr := s.DB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
    // Find or create restaurant by Google Places ID
    result := tx.Where(models.Restaurant{GoogleID: req.Msg.GooglePlacesId}).
        Attrs(models.Restaurant{
            Name:    req.Msg.RestaurantName,
            Address: req.Msg.RestaurantAddress,
            City:    req.Msg.City,
            Country: req.Msg.Country,
        }).
        FirstOrCreate(&restaurant)
    if result.Error != nil {
        return result.Error
    }

    // Remove from wishlist if present (review supersedes wishlist)
    if err := tx.Where("user_id = ? AND restaurant_id = ?", userID, restaurant.ID).
        Delete(&models.WishlistItem{}).Error; err != nil {
        return err
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
```

The `models.WishlistItem` import comes automatically from the `models` package already imported. No additional imports needed.

- [ ] **Step 2: Build and test**

```bash
bunx nx run api:build && cd apps/api && go test ./src/test/... -v
```

Expected: all tests still pass, build succeeds.

- [ ] **Step 3: Commit**

```bash
git add apps/api/src/services/reviews_service.go
git commit -m "feat: auto-remove from wishlist when review is created"
```

---

## Task 6: Frontend — wishlist client + home page buttons

**Files:**
- Modify: `apps/web/src/lib/client/client.ts`
- Modify: `apps/web/src/lib/ui/components/RestaurantSearchSv.svelte`

- [ ] **Step 1: Add wishlist client to client.ts**

In `apps/web/src/lib/client/client.ts`, add the import:

```ts
import { WishlistService } from '$lib/client/generated/wishlist/v1/wishlist_service_pb';
```

Add the client instance:
```ts
const wishlist = createClient(WishlistService, transport);
```

Update the export:
```ts
export default { restaurants, users, googleMaps, auth, reviews, tags, wishlist };
```

- [ ] **Step 2: Add wishlist state to RestaurantSearchSv**

In `apps/web/src/lib/ui/components/RestaurantSearchSv.svelte`, add wishlist state variables alongside the existing review state:

```ts
let isWishlisted = $state(false);
let wishlistLoading = $state(false);
```

- [ ] **Step 3: Check wishlist status when a place is selected**

In `getPlaceDetails`, after the existing review check block, add a wishlist check:

```ts
// Check if current user has this place wishlisted
try {
    const wRes = await clients.wishlist.listWishlist({ googlePlacesId: place.name || '' });
    isWishlisted = (wRes.items?.length ?? 0) > 0;
} catch {
    isWishlisted = false;
}
```

Also reset `isWishlisted = false` at the top of `getPlaceDetails` (where `currentReview = null` is reset):

```ts
selectedPlace = place;
currentReview = null;
isWishlisted = false;
isEditingReview = false;
```

- [ ] **Step 4: Add wishlist handler functions**

After the existing `selectSuggestion` function, add:

```ts
async function toggleWishlist() {
    if (!selectedPlace) return;
    wishlistLoading = true;
    try {
        if (isWishlisted) {
            await clients.wishlist.removeFromWishlist({
                googlePlacesId: selectedPlace.name || ''
            });
            isWishlisted = false;
        } else {
            await clients.wishlist.addToWishlist({
                googlePlacesId: selectedPlace.name || '',
                restaurantName: selectedPlace.displayName?.text || selectedPlace.name || '',
                restaurantAddress: selectedPlace.formattedAddress || '',
                city: extractCity(selectedPlace),
                country: extractCountry(selectedPlace)
            });
            isWishlisted = true;
        }
    } catch (e) {
        console.error('Wishlist toggle error:', e);
    } finally {
        wishlistLoading = false;
    }
}

function extractCity(place: Place): string {
    return place.addressComponents?.find(c => c.types?.includes('locality'))?.longText ?? '';
}

function extractCountry(place: Place): string {
    return place.addressComponents?.find(c => c.types?.includes('country'))?.longText ?? '';
}
```

Note: `Place` is already imported at the top of the script. `addressComponents` and `AddressComponent` are part of the Google Maps proto — check if `types` is the correct field name by looking at the generated TS type for `AddressComponent`. If the field is named differently, adjust accordingly. The `listRestaurantDetails` response includes `addressComponents` when `languageCode` is set.

- [ ] **Step 5: Add wishlist button to the template**

In the template section, find the block that renders `PlacePreviewCard` and the review section. Add a wishlist button row **above** the review section. Replace:

```svelte
{#if selectedPlace}
    <div class="mt-6 space-y-4">
        <PlacePreviewCard place={selectedPlace} />

        {#if isCheckingReview}
```

With:

```svelte
{#if selectedPlace}
    <div class="mt-6 space-y-4">
        <PlacePreviewCard place={selectedPlace} />

        {#if !currentReview}
            <div>
                <Button
                    variant={isWishlisted ? 'outline' : 'secondary'}
                    onclick={toggleWishlist}
                    disabled={wishlistLoading}
                    class="gap-2"
                >
                    {#if wishlistLoading}
                        <div class="h-4 w-4 animate-spin rounded-full border-2 border-current border-t-transparent"></div>
                    {:else if isWishlisted}
                        ★ Wishlisted — click to remove
                    {:else}
                        ☆ Save to wishlist
                    {/if}
                </Button>
            </div>
        {/if}

        {#if isCheckingReview}
```

`Button` is already imported via `import { Button } from '$lib/components/ui/button/index.js'` — add this import if it's not already present.

- [ ] **Step 6: Run svelte-check**

```bash
bunx nx run web:check
```

Expected: 0 errors, 0 warnings.

- [ ] **Step 7: Commit**

```bash
git add apps/web/src/lib/client/client.ts \
        apps/web/src/lib/ui/components/RestaurantSearchSv.svelte
git commit -m "feat: wishlist client + add/remove wishlist buttons on home page"
```

---

## Task 7: Final verification + push

- [ ] **Step 1: Full Go build and test**

```bash
bunx nx run api:build && cd apps/api && go test ./src/test/... -v
```

Expected: build succeeds, all tests pass.

- [ ] **Step 2: Svelte type check**

```bash
bunx nx run web:check
```

Expected: 0 errors.

- [ ] **Step 3: Create PR**

```bash
git push -u origin feat/phase2-wishlist
gh pr create \
  --title "feat: Phase 2 PR2 — Wishlist" \
  --body "Adds Wishlist feature: save/remove restaurants, auto-removed on review creation."
```
