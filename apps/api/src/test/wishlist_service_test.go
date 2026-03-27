package test

import (
	wishlistv1 "api/src/generated/wishlist/v1"
	"api/src/services"
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
