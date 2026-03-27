package test

import (
	reviewsv1 "api/src/generated/reviews/v1"
	"api/src/services"
	"context"
	"errors"
	"testing"

	"connectrpc.com/connect"
)

func TestReviewsService_ListRestaurantReviews_NilDB(t *testing.T) {
	svc := &services.ReviewsService{}
	req := connect.NewRequest(&reviewsv1.ListRestaurantReviewsRequest{
		GooglePlacesId: "places/abc123",
	})
	_, err := svc.ListRestaurantReviews(context.Background(), req)
	if err == nil {
		t.Fatal("expected error from nil DB, got nil")
	}
}

func TestReviewsService_ListRestaurantReviews_MissingGooglePlacesId(t *testing.T) {
	svc := &services.ReviewsService{}
	req := connect.NewRequest(&reviewsv1.ListRestaurantReviewsRequest{})
	_, err := svc.ListRestaurantReviews(context.Background(), req)
	if err == nil {
		t.Fatal("expected error for missing google_places_id, got nil")
	}
}

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
