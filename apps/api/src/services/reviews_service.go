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

	if req.Msg.Id == "" {
		return nil, connect.NewError(connect.CodeInvalidArgument, errors.New("id is required"))
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

	if req.Msg.Id == "" {
		return nil, connect.NewError(connect.CodeInvalidArgument, errors.New("id is required"))
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

	if req.Msg.Id == "" {
		return nil, connect.NewError(connect.CodeInvalidArgument, errors.New("id is required"))
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
